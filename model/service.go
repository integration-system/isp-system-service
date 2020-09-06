package model

import (
	"isp-system-service/entity"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/integration-system/isp-lib/v2/database"
)

var emptyService = (*entity.Service)(nil)

type ServiceRepository struct {
	DB       orm.DB
	rxClient *database.RxDbClient
}

func (r *ServiceRepository) GetServices(list []int32) ([]entity.Service, error) {
	res := make([]entity.Service, 0)
	q := r.getDb().Model(&res)
	if len(res) > 0 {
		q = q.Where("id IN (?)", pg.In(list))
	}
	err := q.Order("created_at DESC").Select()

	return res, err
}

func (r *ServiceRepository) GetServicesByDomainId(domainId ...int32) ([]entity.Service, error) {
	res := make([]entity.Service, 0)
	err := r.getDb().Model(&res).Where("domain_id IN (?)", pg.In(domainId)).Order("created_at DESC").Select()

	return res, err
}

func (r *ServiceRepository) CreateService(service entity.Service) (entity.Service, error) {
	_, err := r.getDb().Model(&service).Insert()

	return service, err
}

func (r *ServiceRepository) GetServiceByNameAndDomainId(name string, domainId int32) (*entity.Service, error) {
	service := new(entity.Service)
	err := r.getDb().Model(service).Where("name = ? AND domain_id = ?", name, domainId).First()
	if err == pg.ErrNoRows {
		return nil, nil
	}

	return service, err
}

func (r *ServiceRepository) UpdateService(service entity.Service) (entity.Service, error) {
	_, err := r.getDb().Model(&service).Column("name", "description").WherePK().Returning("*").Update()

	return service, err
}

func (r *ServiceRepository) GetServiceById(id int32) (*entity.Service, error) {
	service := &entity.Service{Id: id}
	err := r.getDb().Select(service)
	if err == pg.ErrNoRows {
		return nil, nil
	}

	return service, err
}

func (r *ServiceRepository) DeleteServices(list []int32) (int, error) {
	res, err := r.getDb().Model(emptyService).Where("id IN (?)", pg.In(list)).Delete()
	if err != nil {
		return 0, err
	}

	return res.RowsAffected(), nil
}

func (r *ServiceRepository) getDb() orm.DB {
	if r.DB != nil {
		return r.DB
	}

	return r.rxClient.Unsafe()
}
