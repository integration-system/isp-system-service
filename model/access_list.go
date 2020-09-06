package model

import (
	"isp-system-service/entity"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/integration-system/isp-lib/v2/database"
)

type AccessListRepository struct {
	DB       orm.DB
	rxClient *database.RxDbClient
}

func (r *AccessListRepository) GetByAppId(appId int32) ([]entity.AccessList, error) {
	model := make([]entity.AccessList, 0)
	if err := r.getDb().Model(&model).Where("app_id = ?", appId).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return model, nil
}

func (r *AccessListRepository) GetByAppIdList(list []int32) ([]entity.AccessList, error) {
	model := make([]entity.AccessList, 0)
	if err := r.getDb().Model(&model).Where("app_id IN (?)", pg.In(list)).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return model, nil
}

func (r *AccessListRepository) DeleteById(id int32) ([]entity.AccessList, error) {
	model := make([]entity.AccessList, 0)
	if _, err := r.getDb().Model(&model).Where("app_id = ?", id).Returning("*").Delete(); err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}

		return nil, err
	} else {
		return model, nil
	}
}

func (r *AccessListRepository) DeleteByIdList(list []int32) ([]entity.AccessList, error) {
	model := make([]entity.AccessList, 0)
	if _, err := r.getDb().Model(&model).Where("app_id IN (?)", pg.In(list)).Returning("*").Delete(); err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}

		return nil, err
	} else {
		return model, nil
	}
}

func (r *AccessListRepository) UpsertArray(model []entity.AccessList) (int, error) {
	_, _ = r.getDb().Model(&model).WherePK().Delete()
	if result, err := r.getDb().Model(&model).Insert(); err != nil {
		return 0, err
	} else {
		return result.RowsAffected(), nil
	}
}

func (r *AccessListRepository) Upsert(model entity.AccessList) (int, error) {
	if result, err := r.getDb().Model(&model).OnConflict("(app_id, method) DO UPDATE").Insert(); err != nil {
		return 0, err
	} else {
		return result.RowsAffected(), nil
	}
}

func (r *AccessListRepository) getDb() orm.DB {
	if r.DB != nil {
		return r.DB
	}

	return r.rxClient.Unsafe()
}
