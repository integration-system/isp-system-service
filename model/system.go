package model

import (
	"isp-system-service/entity"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/integration-system/isp-lib/v2/database"
)

var emptySystem = (*entity.System)(nil)

type SystemRepository struct {
	DB       orm.DB
	rxClient *database.RxDbClient
}

func (r *SystemRepository) GetSystems(list []int32) ([]entity.System, error) {
	res := make([]entity.System, 0)
	q := r.getDb().Model(&res)
	if len(res) > 0 {
		q = q.Where("id IN (?)", pg.In(list))
	}
	err := q.Order("created_at DESC").Select()

	return res, err
}

func (r *SystemRepository) CreateSystem(system entity.System) (entity.System, error) {
	_, err := r.getDb().Model(&system).Insert()

	return system, err
}

func (r *SystemRepository) GetSystemByName(name string) (*entity.System, error) {
	sys := new(entity.System)
	err := r.getDb().Model(sys).Where("name = ?", name).First()
	if err == pg.ErrNoRows {
		return nil, nil
	}

	return sys, err
}

func (r *SystemRepository) UpdateSystem(system entity.System) (entity.System, error) {
	_, err := r.getDb().Model(&system).Column("name", "description").WherePK().Returning("*").Update()

	return system, err
}

func (r *SystemRepository) GetSystemById(id int32) (*entity.System, error) {
	sys := &entity.System{Id: id}
	err := r.getDb().Select(sys)
	if err == pg.ErrNoRows {
		return nil, nil
	}

	return sys, err
}

func (r *SystemRepository) DeleteSystems(list []int32) (int, error) {
	res, err := r.getDb().Model(emptySystem).Where("id IN (?)", pg.In(list)).Delete()
	if err != nil {
		return 0, err
	}

	return res.RowsAffected(), nil
}

func (r *SystemRepository) getDb() orm.DB {
	if r.DB != nil {
		return r.DB
	}

	return r.rxClient.Unsafe()
}
