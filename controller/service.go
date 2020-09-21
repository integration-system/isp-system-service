package controller

import (
	"isp-system-service/domain"
	"isp-system-service/entity"
	"isp-system-service/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var Service serviceController

type serviceController struct{}

// GetService godoc
// @Tags service
// @Summary Получить список сервисов
// @Description Возвращает список сервисов по их идентификаторам
// @Accept  json
// @Produce  json
// @Param body body []integer false "Массив идентификаторов сервисов"
// @Success 200 {array} entity.Service
// @Failure 500 {object} structure.GrpcError
// @Router /service/get_service [POST]
func (serviceController) GetService(list []int32) ([]entity.Service, error) {
	res, err := model.ServiceRep.GetServices(list)
	if err != nil {
		return res, err
	}

	return res, nil
}

// GetServicesByDomainId godoc
// @Tags service
// @Summary Получить список сервисов по идентификатору домена
// @Description Возвращает список сервисов по идентификатору домена
// @Accept  json
// @Produce  json
// @Param body body domain.Identity true "Идентификатор домена"
// @Success 200 {array} entity.Service
// @Failure 500 {object} structure.GrpcError
// @Router /service/get_services_by_domain_id [POST]
func (serviceController) GetServicesByDomainId(identity domain.Identity) ([]entity.Service, error) {
	return model.ServiceRep.GetServicesByDomainId(identity.Id)
}

// CreateUpdateService godoc
// @Tags service
// @Summary Создать/обновить сервис
// @Description Если сервис с такими идентификатором существует, то обновляет данные, если нет, то добавляет данные в базу
// @Accept  json
// @Produce  json
// @Param body body entity.Service true "Объект сервиса"
// @Success 200 {object} entity.Service
// @Failure 400 {object} structure.GrpcError
// @Failure 404 {object} structure.GrpcError
// @Failure 409 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /service/create_update_service [POST]
func (serviceController) CreateUpdateService(service entity.Service) (*entity.Service, error) {
	existed, err := model.ServiceRep.GetServiceByNameAndDomainId(service.Name, service.DomainId)
	if err != nil {
		return nil, err
	}

	d, err := model.DomainRep.GetDomainById(service.DomainId)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Domain with id %d not found", service.DomainId)
	}

	if service.Id == 0 {
		if existed != nil {
			return nil, status.Errorf(codes.AlreadyExists, "Service with name %s already exists", service.Name)
		}
		service, err = model.ServiceRep.CreateService(service)
		if err != nil {
			return nil, err
		}

		return &service, nil
	}

	if existed != nil && existed.Id != service.Id {
		return nil, status.Errorf(codes.AlreadyExists, "Service with name %s already exists", service.Name)
	}

	existed, err = model.ServiceRep.GetServiceById(service.Id)
	if err != nil {
		return nil, err
	}
	if existed == nil {
		return nil, status.Errorf(codes.NotFound, "Service with id %d not found", service.Id)
	}

	service, err = model.ServiceRep.UpdateService(service)
	if err != nil {
		return nil, err
	}

	return &service, nil
}

// GetServiceById godoc
// @Tags service
// @Summary Получить сервис по идентификатору
// @Description Возвращает описание сервиса по его идентификатору
// @Accept  json
// @Produce  json
// @Param body body domain.Identity true "Идентификатор сервиса"
// @Success 200 {object} entity.Service
// @Failure 404 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /service/get_service_by_id [POST]
func (serviceController) GetServiceById(identity domain.Identity) (*entity.Service, error) {
	service, err := model.ServiceRep.GetServiceById(identity.Id)
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, status.Errorf(codes.NotFound, "Service with id %d not found", identity.Id)
	}

	return service, nil
}

// DeleteService godoc
// @Tags service
// @Summary Удалить сервисы
// @Description Удаляет сервисов по списку их идентификаторов, возвращает количество удаленных сервисов
// @Accept  json
// @Produce  json
// @Param body body []integer true "Массив идентификаторов сервисов"
// @Success 200 {object} domain.DeleteResponse
// @Failure 400 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /service/delete_service [POST]
func (serviceController) DeleteService(list []int32) (domain.DeleteResponse, error) {
	if len(list) == 0 {
		return domain.DeleteResponse{Deleted: 0}, status.Error(codes.InvalidArgument, "At least one id are required")
	}

	res, err := model.ServiceRep.DeleteServices(list)
	if err != nil {
		return domain.DeleteResponse{}, err
	}

	return domain.DeleteResponse{Deleted: res}, nil
}
