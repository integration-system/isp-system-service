package model

import (
	"isp-system-service/entity"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/integration-system/isp-lib/v2/database"
)

var emptyDomain = (*entity.Domain)(nil)

type DomainRepository struct {
	DB       orm.DB
	rxClient *database.RxDbClient
}

func (r *DomainRepository) GetDomains(list []int32) ([]entity.Domain, error) {
	res := make([]entity.Domain, 0)
	q := r.getDb().Model(&res)
	if len(res) > 0 {
		q = q.Where("id IN (?)", pg.In(list))
	}
	err := q.Order("created_at DESC").Select()

	return res, err
}

func (r *DomainRepository) GetDomainsBySystemId(systemId int32) ([]entity.Domain, error) {
	res := make([]entity.Domain, 0)
	err := r.getDb().Model(&res).Where("system_id = ?", systemId).Order("created_at DESC").Select()

	return res, err
}

func (r *DomainRepository) CreateDomain(domain entity.Domain) (entity.Domain, error) {
	_, err := r.getDb().Model(&domain).Insert()

	return domain, err
}

func (r *DomainRepository) GetDomainByNameAndSystemId(name string, systemId int32) (*entity.Domain, error) {
	domain := new(entity.Domain)
	err := r.getDb().Model(domain).Where("name = ? AND system_id = ?", name, systemId).First()
	if err == pg.ErrNoRows {
		return nil, nil
	}

	return domain, err
}

func (r *DomainRepository) UpdateDomain(domain entity.Domain) (entity.Domain, error) {
	_, err := r.getDb().Model(&domain).Column("name", "description").WherePK().Returning("*").Update()

	return domain, err
}

func (r *DomainRepository) GetDomainById(id int32) (*entity.Domain, error) {
	domain := &entity.Domain{Id: id}
	err := r.getDb().Select(domain)
	if err == pg.ErrNoRows {
		return nil, nil
	}

	return domain, err
}

func (r *DomainRepository) DeleteDomains(list []int32) (int, error) {
	res, err := r.getDb().Model(emptyDomain).Where("id IN (?)", pg.In(list)).Delete()
	if err != nil {
		return 0, err
	}

	return res.RowsAffected(), nil
}

func (r *DomainRepository) getDb() orm.DB {
	if r.DB != nil {
		return r.DB
	}

	return r.rxClient.Unsafe()
}
