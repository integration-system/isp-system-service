package controller

import (
	"context"
	"fmt"

	"isp-system-service/domain"
	"isp-system-service/entity"
	"isp-system-service/model"
	"isp-system-service/redis"

	rd "github.com/go-redis/redis/v8"
	rdLib "github.com/integration-system/isp-lib/v2/redis"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var AccessList accessListController

type accessListController struct{}

// GetById godoc
// @Tags accessList
// @Summary Получить список доступности методов для приложения
// @Description Возвращает список методов для приложения, для которых заданы настройки доступа
// @Accept  json
// @Produce  json
// @Param body body domain.Identity false "идентификатор приложения"
// @Success 200 {array} domain.MethodInfo "список доступности методов"
// @Failure 404 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /access_list/get_by_id [POST]
func (c accessListController) GetById(request domain.Identity) ([]domain.MethodInfo, error) {
	err := c.checkAppById(request.Id)
	if err != nil {
		return nil, err
	}

	return c.getAccessListById(request.Id)
}

// SetOne godoc
// @Tags accessList
// @Summary Настроить доступность метода для приложения
// @Description Возвращает количество изменных строк
// @Accept  json
// @Produce  json
// @Param body body entity.AccessList false "объект для настройки доступа"
// @Success 200 {object} domain.CountResponse "количество измененных строк"
// @Failure 404 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /access_list/set_one [POST]
func (c accessListController) SetOne(request entity.AccessList) (*domain.CountResponse, error) {
	err := c.checkAppById(request.AppId)
	if err != nil {
		return nil, err
	}

	resp := 0
	err = model.DbClient.RunInTransaction(func(repository model.AccessListRepository) error {
		resp, err = repository.Upsert(request)
		if err != nil {
			return err
		}

		_, err = redis.Client.Get().UseDb(rdLib.ApplicationPermissionDb, func(p rd.Pipeliner) error {
			key := fmt.Sprintf("%d|%s", request.AppId, request.Method)
			_, err := p.Set(context.Background(), key, request.Value, 0).Result()

			return err
		})

		return err
	})
	if err != nil {
		return nil, err
	}

	return &domain.CountResponse{Count: resp}, nil
}

// SetList godoc
// @Tags accessList
// @Summary Настройть доступность списка методов для приложения
// @Description Возвращает список методов для приложения, для которых заданы настройки доступа
// @Accept  json
// @Produce  json
// @Param body body domain.SetListRequest false "объект настройки доступа"
// @Success 200 {array} domain.MethodInfo "список доступности методов"
// @Failure 404 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /access_list/set_list [POST]
func (c accessListController) SetList(request domain.SetListRequest) ([]domain.MethodInfo, error) {
	err := c.checkAppById(request.AppId)
	if err != nil {
		return nil, err
	}

	oldAccessList := make([]entity.AccessList, 0)
	err = model.DbClient.RunInTransaction(func(repository model.AccessListRepository) error {
		if request.RemoveOld {
			oldAccessList, err = repository.DeleteById(request.AppId)
			if err != nil {
				return err
			}
		}

		newAccessList := make([]entity.AccessList, len(request.Methods))
		redisMsetRequest := make([]interface{}, 0)
		for i, m := range request.Methods {
			newAccessList[i] = entity.AccessList{
				AppId:  request.AppId,
				Method: m.Method,
				Value:  m.Value,
			}
			redisMsetRequest = append(redisMsetRequest, fmt.Sprintf("%d|%s", request.AppId, m.Method), m.Value)
		}

		redisDelRequest := make([]string, len(oldAccessList))
		for i, access := range oldAccessList {
			redisDelRequest[i] = fmt.Sprintf("%d|%s", request.AppId, access.Method)
		}

		if len(newAccessList) > 0 {
			_, err := repository.UpsertArray(newAccessList)
			if err != nil {
				return err
			}
		}

		_, err := redis.Client.Get().UseDb(rdLib.ApplicationPermissionDb, func(p rd.Pipeliner) error {
			if len(redisDelRequest) > 0 {
				_, err := p.Del(context.Background(), redisDelRequest...).Result()
				if err != nil {
					return err
				}
			}

			if len(redisMsetRequest) > 0 {
				_, err := p.MSet(context.Background(), redisMsetRequest...).Result()
				if err != nil {
					return err
				}
			}

			return nil
		})

		return err
	})
	if err != nil {
		return nil, err
	}

	return c.getAccessListById(request.AppId)
}

func (c accessListController) getAccessListById(id int32) ([]domain.MethodInfo, error) {
	accessList, err := model.AccessListRep.GetByAppId(id)
	if err != nil {
		return nil, err
	}

	return c.convertAccessList(accessList), nil
}

func (accessListController) convertAccessList(accessLists []entity.AccessList) []domain.MethodInfo {
	methodInfos := make([]domain.MethodInfo, len(accessLists))
	for i, access := range accessLists {
		methodInfos[i] = domain.MethodInfo{
			Method: access.Method,
			Value:  access.Value,
		}
	}

	return methodInfos
}

func (accessListController) checkAppById(appId int32) error {
	app, err := model.AppRep.GetApplicationById(appId)
	if err != nil {
		return err
	}
	if app == nil {
		return status.Errorf(codes.NotFound, "application '%d' not found", appId)
	}

	return nil
}
