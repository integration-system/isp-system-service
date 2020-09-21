package controller

import (
	"context"
	"fmt"

	"isp-system-service/conf"
	"isp-system-service/domain"
	"isp-system-service/entity"
	"isp-system-service/model"
	"isp-system-service/redis"

	rd "github.com/go-redis/redis/v8"
	"github.com/integration-system/isp-lib/v2/config"
	rdLib "github.com/integration-system/isp-lib/v2/redis"
	_ "github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var Application applicationController

type applicationController struct{}

// GetApplications godoc
// @Tags application
// @Summary Получить список приложений
// @Description Возвращает массив приложений с токенами по их идентификаторам
// @Accept  json
// @Produce  json
// @Param body body []integer false "Массив идентификаторов приложений"
// @Success 200 {array} domain.AppWithToken
// @Failure 500 {object} structure.GrpcError
// @Router /application/get_applications [POST]
func (c applicationController) GetApplications(list []int32) ([]*domain.AppWithToken, error) {
	res, err := model.AppRep.GetApplications(list)
	if err != nil {
		return nil, err
	}

	return c.enrichWithTokens(res...)
}

// GetApplicationsByServiceId godoc
// @Tags application
// @Summary Получить список приложений по идентификатору сервиса
// @Description Возвращает список приложений по запрошенныму идентификатору сервиса
// @Accept  json
// @Produce  json
// @Param body body domain.Identity true "Идентификатор серсиса"
// @Success 200 {array} domain.AppWithToken
// @Failure 500 {object} structure.GrpcError
// @Router /application/get_applications_by_service_id [POST]
func (c applicationController) GetApplicationsByServiceId(identity domain.Identity) ([]*domain.AppWithToken, error) {
	arr, err := model.AppRep.GetApplicationsByServiceId(identity.Id)
	if err != nil {
		return nil, err
	}

	return c.enrichWithTokens(arr...)
}

// CreateUpdateApplication godoc
// @Tags application
// @Summary Создать/обновить приложение
// @Description Если приложение с такими идентификатором существует, то обновляет данные, если нет, то добавляет данные в базу
// @Accept  json
// @Produce  json
// @Param body body entity.Application true "Объект приложения"
// @Success 200 {object} domain.AppWithToken
// @Failure 400 {object} structure.GrpcError
// @Failure 404 {object} structure.GrpcError
// @Failure 409 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /application/create_update_application [POST]
func (c applicationController) CreateUpdateApplication(app entity.Application) (*domain.AppWithToken, error) {
	existed, err := model.AppRep.GetApplicationByNameAndServiceId(app.Name, app.ServiceId)
	if err != nil {
		return nil, err
	}

	s, err := model.ServiceRep.GetServiceById(app.ServiceId)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Service with id %d not found", app.ServiceId)
	}

	if app.Id == 0 {
		if existed != nil {
			return nil, status.Errorf(codes.AlreadyExists, "Application with name %s already exists", app.Name)
		}

		app, err = model.AppRep.CreateApplication(app)
		if err != nil {
			return nil, err
		}

		_, err = c.enrichWithTokens(app)
		if err != nil {
			return nil, err
		}

		return &domain.AppWithToken{App: app, Tokens: []entity.Token{}}, nil
	}

	if existed != nil && existed.Id != app.Id {
		return nil, status.Errorf(codes.AlreadyExists, "Application with name %s already exists", app.Name)
	}

	existed, err = model.AppRep.GetApplicationById(app.Id)
	if err != nil {
		return nil, err
	}
	if existed == nil {
		return nil, status.Errorf(codes.NotFound, "Application with id %d not found", app.Id)
	}

	app, err = model.AppRep.UpdateApplication(app)
	if err != nil {
		return nil, err
	}
	arr, err := c.enrichWithTokens(app)
	if err != nil {
		return nil, err
	}

	return arr[0], nil
}

// GetApplicationById godoc
// @Tags application
// @Summary Получить приложение по идентификатору
// @Description  Возвращает описание приложения по его идентификатору
// @Accept  json
// @Produce  json
// @Param body body domain.Identity true "Идентификатор приложения"
// @Success 200 {object} domain.AppWithToken
// @Failure 400 {object} structure.GrpcError
// @Failure 404 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /application/get_application_by_id [POST]
func (c applicationController) GetApplicationById(identity domain.Identity) (*domain.AppWithToken, error) {
	application, err := model.AppRep.GetApplicationById(identity.Id)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, status.Errorf(codes.NotFound, "Application with id %d not found", identity.Id)
	}

	arr, err := c.enrichWithTokens(*application)
	if err != nil {
		return nil, err
	}

	return arr[0], nil
}

// DeleteApplications godoc
// @Tags application
// @Summary Удалить приложения
// @Description Удаляет приложения по списку их идентификаторов, возвращает количество удаленных приложений
// @Accept  json
// @Produce  json
// @Param body body []integer false "Массив идентификаторов приложений"
// @Success 200 {object} domain.DeleteResponse
// @Failure 400 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /application/delete_applications [POST]
func (applicationController) DeleteApplications(list []int32) (domain.DeleteResponse, error) {
	if len(list) == 0 {
		return domain.DeleteResponse{}, status.Errorf(codes.InvalidArgument, "At least one id are required")
	}

	var (
		count        = 0
		instanceUuid = config.Get().(*conf.Configuration).InstanceUuid
	)
	_, err := redis.Client.Get().UseDb(rdLib.ApplicationTokenDb, func(p rd.Pipeliner) error {
		return model.DbClient.RunInTransaction(func(
			appRep model.AppRepository, tokenRep model.TokenRepository, accessRep model.AccessListRepository) error {
			accessList, err := accessRep.GetByAppIdList(list)
			if err != nil {
				return err
			}

			redisDelRequest := make([]string, len(accessList))
			for i, access := range accessList {
				redisDelRequest[i] = fmt.Sprintf("%d|%s", access.AppId, access.Method)
			}

			tokens, err := tokenRep.GetTokensByAppId(list...)
			if err != nil {
				return err
			}

			tokenIdList := make([]string, len(tokens))
			for i, t := range tokens {
				tokenIdList[i] = t.Token
			}

			if len(tokenIdList) != 0 {
				keys := make([]string, len(tokens))
				for i, token := range tokenIdList {
					keys[i] = fmt.Sprintf("%s|%s", token, instanceUuid)
				}

				_, err := p.Del(context.Background(), keys...).Result()
				if err != nil {
					return err
				}
			}

			count, err = appRep.DeleteApplications(list)
			if err != nil {
				return err
			}

			if len(redisDelRequest) > 0 {
				_, err := redis.Client.Get().UseDb(rdLib.ApplicationPermissionDb, func(p rd.Pipeliner) error {
					_, err := p.Del(context.Background(), redisDelRequest...).Result()

					return err
				})
				if err != nil {
					return err
				}
			}

			return nil
		})
	})
	if err != nil {
		return domain.DeleteResponse{}, err
	}

	return domain.DeleteResponse{Deleted: count}, nil
}

// GetSystemTree godoc
// @Tags application
// @Summary Метод получения системного дерева
// @Description Возвращает описание взаимосвязей сервисов и приложений
// @Accept  json
// @Produce  json
// @Success 200 {array} domain.DomainWithServices
// @Failure 500 {object} structure.GrpcError
// @Router /application/get_system_tree [POST]
func (applicationController) GetSystemTree(md metadata.MD) ([]*domain.DomainWithServices, error) {
	sysId, err := utils.ResolveMetadataIdentity(utils.SystemIdHeader, md)
	if err != nil {
		return nil, err
	}
	domains, err := model.DomainRep.GetDomainsBySystemId(int32(sysId))
	if err != nil {
		return nil, err
	}

	l := len(domains)
	if l == 0 {
		return []*domain.DomainWithServices{}, nil
	}
	idList := make([]int32, l)
	domainsMap := make(map[int32]*domain.DomainWithServices, l)
	res := make([]*domain.DomainWithServices, l)
	for i, d := range domains {
		idList[i] = d.Id
		dws := &domain.DomainWithServices{Id: d.Id, Name: d.Name, Description: d.Description, Services: make([]*domain.ServiceWithApps, 0)}
		domainsMap[d.Id] = dws
		res[i] = dws
	}
	services, err := model.ServiceRep.GetServicesByDomainId(idList...)
	if err != nil {
		return nil, err
	}

	l = len(services)
	idList = make([]int32, l)
	servicesMap := make(map[int32]*domain.ServiceWithApps, l)
	for i, s := range services {
		idList[i] = s.Id
		d := domainsMap[s.DomainId]
		swa := &domain.ServiceWithApps{Id: s.Id, Name: s.Name, Description: s.Description, Apps: make([]*domain.SimpleApp, 0)}
		d.Services = append(d.Services, swa)
		servicesMap[s.Id] = swa
	}
	apps, err := model.AppRep.GetApplicationsByServiceId(idList...)
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		s := servicesMap[app.ServiceId]
		s.Apps = append(s.Apps, &domain.SimpleApp{
			Id:          app.Id,
			Name:        app.Name,
			Type:        app.Type,
			Description: app.Description,
			Tokens:      make([]entity.Token, 0),
		})
	}

	return res, nil
}

func (applicationController) enrichWithTokens(apps ...entity.Application) ([]*domain.AppWithToken, error) {
	l := len(apps)
	if len(apps) == 0 {
		return []*domain.AppWithToken{}, nil
	}
	idList := make([]int32, l)
	appMap := make(map[int32]*domain.AppWithToken, l)
	enriched := make([]*domain.AppWithToken, l)
	for i, a := range apps {
		idList[i] = a.Id
		awt := &domain.AppWithToken{App: a, Tokens: make([]entity.Token, 0)}
		appMap[a.Id] = awt
		enriched[i] = awt
	}
	tokens, err := model.TokenRep.GetTokensByAppId(idList...)
	if err != nil {
		return nil, err
	}

	for _, v := range tokens {
		app := appMap[v.AppId]
		app.Tokens = append(app.Tokens, v)
	}

	return enriched, nil
}
