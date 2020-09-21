package controller

import (
	"context"
	"fmt"
	"time"

	rd "github.com/go-redis/redis/v8"
	"github.com/integration-system/isp-lib/v2/config"
	redisLib "github.com/integration-system/isp-lib/v2/redis"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"isp-system-service/conf"
	"isp-system-service/domain"
	"isp-system-service/entity"
	"isp-system-service/model"
	"isp-system-service/redis"
	"isp-system-service/service"
)

const (
	SystemIdentityFieldInDb      = "1"
	DomainIdentityFieldInDb      = "2"
	ServiceIdentityFieldInDb     = "3"
	ApplicationIdentityFieldInDb = "4"
)

var Token tokenController

type tokenController struct{}

// CreateToken godoc
// @Tags token
// @Summary Создать токен
// @Description Созддает токен и привязывает его к приложению
// @Accept  json
// @Produce  json
// @Param body body domain.CreateTokenRequest true "Объект создания токена"
// @Success 200 {object} domain.AppWithToken
// @Failure 500 {object} structure.GrpcError
// @Router /token/create_token [POST]
func (c tokenController) CreateToken(req domain.CreateTokenRequest) (*domain.AppWithToken, error) {
	m, err, app := c.getIdMap(req.AppId)
	if err != nil {
		return nil, err
	}

	expTime := req.ExpireTimeMs
	if expTime == 0 {
		expTime = config.GetRemote().(*conf.RemoteConfig).DefaultTokenExpireTime
	}
	token, err := service.Jwt.CreateApplication(req.AppId, expTime)
	if err != nil {
		return nil, err
	}

	tokenInfo := entity.Token{
		Token:      token,
		ExpireTime: expTime,
		AppId:      req.AppId,
		CreatedAt:  time.Now(),
	}

	if err := model.DbClient.RunInTransaction(func(repository model.TokenRepository) error {
		tokenInfo, err = repository.SaveToken(tokenInfo)
		if err != nil {
			return err
		}

		return c.setIdentityMapForToken(tokenInfo, m)
	}); err != nil {
		return nil, err
	}

	arr, err := Application.enrichWithTokens(*app)
	if err != nil {
		return nil, err
	}

	return arr[0], nil
}

// RevokeTokens godoc
// @Tags token
// @Summary Отозвать токены
// @Description Отвязывает токены от приложений и удялет их
// @Accept  json
// @Produce  json
// @Param body body domain.RevokeTokensRequest true "Объект отзыва токенов"
// @Success 200 {object} domain.AppWithToken
// @Failure 500 {object} structure.GrpcError
// @Router /token/revoke_tokens [POST]
func (c tokenController) RevokeTokens(req domain.RevokeTokensRequest) (*domain.AppWithToken, error) {
	app, err := model.AppRep.GetApplicationById(req.AppId)
	if err != nil {
		return nil, err
	}

	_, err = c.revokeTokens(req.Tokens)
	if err != nil {
		return nil, err
	}

	res, err := Application.enrichWithTokens(*app)
	if err != nil {
		return nil, err
	}

	return res[0], nil
}

// RevokeTokensForApp godoc
// @Tags token
// @Summary Отозвать токены для приложения
// @Description Отвязывает токены от приложений и удаляет их по идентификатору приложения
// @Accept  json
// @Produce  json
// @Param body body domain.Identity true "Идентификатор приложения"
// @Success 200 {object} domain.DeleteResponse
// @Failure 500 {object} structure.GrpcError
// @Router /token/revoke_tokens_for_app [POST]
func (c tokenController) RevokeTokensForApp(identity domain.Identity) (*domain.DeleteResponse, error) {
	return c.revokeTokensForApp(identity)
}

// GetTokensByAppId godoc
// @Tags token
// @Summary Получить токены по идентификаотру приложения
// @Description Возвращает список токенов, привязанных к приложению
// @Accept  json
// @Produce  json
// @Param body body domain.Identity true "Идентификатор приложения"
// @Success 200 {array} entity.Token
// @Failure 500 {object} structure.GrpcError
// @Router /token/get_tokens_by_app_id [POST]
func (tokenController) GetTokensByAppId(identity domain.Identity) ([]entity.Token, error) {
	return model.TokenRep.GetTokensByAppId(identity.Id)
}

func (c tokenController) setIdentityMapForToken(token entity.Token, idMap map[string]interface{}) error {
	return c.SetIdentityMapForTokenV2(token.Token, token.ExpireTime, idMap)
}

func (tokenController) SetIdentityMapForTokenV2(token string, expireTime int64, idMap map[string]interface{}) error {
	instanceUuid := config.Get().(*conf.Configuration).InstanceUuid
	_, e := redis.Client.Get().UseDbTx(redisLib.ApplicationTokenDb, func(p rd.Pipeliner) error {
		key := fmt.Sprintf("%s|%s", token, instanceUuid)
		stat := p.HMSet(context.Background(), key, idMap)
		err := stat.Err()
		if err != nil {
			return err
		}
		if expireTime > 0 {
			err = p.Expire(context.Background(), token, time.Duration(expireTime)*time.Millisecond).Err()
		}

		return err
	})

	return e
}

func (tokenController) getIdMap(appId int32) (map[string]interface{}, error, *entity.Application) {
	appInfo, err := model.AppRep.GetApplicationById(appId)
	if err != nil {
		return nil, err, nil
	}
	if appInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Application with id %d not found", appId), nil
	}

	serviceInfo, err := model.ServiceRep.GetServiceById(appInfo.ServiceId)
	if err != nil {
		return nil, err, appInfo
	}
	if serviceInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Service with id %d not found", appInfo.ServiceId), appInfo
	}

	domainInfo, err := model.DomainRep.GetDomainById(serviceInfo.DomainId)
	if err != nil {
		return nil, err, appInfo
	}
	if domainInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Domain with id %d not found", serviceInfo.DomainId), appInfo
	}

	systemInfo, err := model.SystemRep.GetSystemById(domainInfo.SystemId)
	if err != nil {
		return nil, err, appInfo
	}
	if systemInfo == nil {
		return nil, status.Errorf(codes.NotFound, "System with id %d not found", domainInfo.SystemId), appInfo
	}

	return map[string]interface{}{
		SystemIdentityFieldInDb:      systemInfo.Id,
		DomainIdentityFieldInDb:      domainInfo.Id,
		ServiceIdentityFieldInDb:     serviceInfo.Id,
		ApplicationIdentityFieldInDb: appInfo.Id,
	}, nil, appInfo
}

func (c tokenController) revokeTokensForApp(identity domain.Identity) (*domain.DeleteResponse, error) {
	tokens, err := model.TokenRep.GetTokensByAppId(identity.Id)
	if err != nil {
		return nil, err
	}

	tokenIdList := make([]string, len(tokens))
	for i, t := range tokens {
		tokenIdList[i] = t.Token
	}

	return c.revokeTokens(tokenIdList)
}

func (tokenController) revokeTokens(tokens []string) (*domain.DeleteResponse, error) {
	if len(tokens) == 0 {
		return &domain.DeleteResponse{Deleted: 0}, nil
	}

	var (
		count        = 0
		instanceUuid = config.Get().(*conf.Configuration).InstanceUuid
		keys         = make([]string, len(tokens))
	)
	for i, token := range tokens {
		keys[i] = fmt.Sprintf("%s|%s", token, instanceUuid)
	}
	err := model.DbClient.RunInTransaction(func(rep model.TokenRepository) error {
		deleted, err := rep.DeleteTokens(tokens)
		if err != nil {
			return err
		}
		count = deleted

		_, err = redis.Client.Get().UseDbTx(redisLib.ApplicationTokenDb, func(p rd.Pipeliner) error {
			_, err := p.Del(context.Background(), keys...).Result()

			return err
		})

		return err
	})
	if err != nil {
		return nil, err
	}

	return &domain.DeleteResponse{Deleted: count}, nil
}
