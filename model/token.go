package model

import (
	"isp-system-service/entity"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/integration-system/isp-lib/v2/database"
)

var emptyToken = (*entity.Token)(nil)

type TokenRepository struct {
	DB       orm.DB
	rxClient *database.RxDbClient
}

func (r *TokenRepository) SaveToken(token entity.Token) (entity.Token, error) {
	_, err := r.getDb().Model(&token).Insert()

	return token, err
}

func (r *TokenRepository) GetTokenById(id string) (*entity.Token, error) {
	token := &entity.Token{Token: id}
	err := r.getDb().Select(token)
	if err == pg.ErrNoRows {
		return nil, nil
	}

	return token, err
}

func (r *TokenRepository) GetTokensByAppId(appId ...int32) ([]entity.Token, error) {
	res := make([]entity.Token, 0)
	err := r.getDb().Model(&res).Where("app_id IN (?)", pg.In(appId)).Order("created_at DESC").Select()

	return res, err
}

func (r *TokenRepository) DeleteTokens(tokens []string) (int, error) {
	res, err := r.getDb().Model(emptyToken).Where("token IN (?)", pg.In(tokens)).Delete()
	if err != nil {
		return 0, err
	}

	return res.RowsAffected(), nil
}

func (r *TokenRepository) getDb() orm.DB {
	if r.DB != nil {
		return r.DB
	}

	return r.rxClient.Unsafe()
}
