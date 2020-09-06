package model

import (
	"isp-system-service/entity"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/integration-system/isp-lib/v2/database"
)

var emptyApplication = (*entity.Application)(nil)

type AppRepository struct {
	DB       orm.DB
	rxClient *database.RxDbClient
}

func (r *AppRepository) GetApplications(list []int32) ([]entity.Application, error) {
	res := make([]entity.Application, 0)
	q := r.getDb().Model(&res)
	if len(res) > 0 {
		q = q.Where("id IN (?)", pg.In(list))
	}
	err := q.Order("created_at DESC").Select()

	return res, err
}

func (r *AppRepository) GetApplicationsByServiceId(serviceId ...int32) ([]entity.Application, error) {
	res := make([]entity.Application, 0)
	err := r.getDb().Model(&res).Where("service_id IN (?)", pg.In(serviceId)).Order("created_at DESC").Select()

	return res, err
}

func (r *AppRepository) CreateApplication(service entity.Application) (entity.Application, error) {
	_, err := r.getDb().Model(&service).Insert()

	return service, err
}

func (r *AppRepository) GetApplicationByNameAndServiceId(name string, serviceId int32) (*entity.Application, error) {
	app := new(entity.Application)
	err := r.getDb().Model(app).Where("name = ? AND service_id = ?", name, serviceId).First()
	if err == pg.ErrNoRows {
		return nil, nil
	}

	return app, err
}

func (r *AppRepository) UpdateApplication(app entity.Application) (entity.Application, error) {
	_, err := r.getDb().Model(&app).Column("name", "description").WherePK().Returning("*").Update()

	return app, err
}

func (r *AppRepository) GetApplicationById(id int32) (*entity.Application, error) {
	app := &entity.Application{Id: id}
	err := r.getDb().Select(app)
	if err == pg.ErrNoRows {
		return nil, nil
	}

	return app, err
}

func (r *AppRepository) DeleteApplications(list []int32) (int, error) {
	res, err := r.getDb().Model(emptyApplication).Where("id IN (?)", pg.In(list)).Delete()
	if err != nil {
		return 0, err
	}

	return res.RowsAffected(), nil
}

func (r *AppRepository) getDb() orm.DB {
	if r.DB != nil {
		return r.DB
	}

	return r.rxClient.Unsafe()
}
