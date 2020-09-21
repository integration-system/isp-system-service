package controller

import (
	"fmt"

	"isp-system-service/domain"
	"isp-system-service/entity"
	"isp-system-service/model"

	_ "github.com/integration-system/isp-lib/v2/structure"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var System systemController

type systemController struct{}

// GetSystems godoc
// @Tags system
// @Summary Получить список систем
// @Description Возвращает список систем по их идентификаторам
// @Accept  json
// @Produce  json
// @Param body body []integer false "Массив идентификаторов систем"
// @Success 200 {array} entity.System
// @Failure 500 {object} structure.GrpcError
// @Router /system/get_systems [POST]
func (systemController) GetSystems(list []int32) ([]entity.System, error) {
	res, err := model.SystemRep.GetSystems(list)
	if err != nil {
		return res, err
	}

	return res, nil
}

// CreateUpdateSystem godoc
// @Tags system
// @Summary Создать/обновить систему
// @Description Если система с такими идентификатором существует, то обновляет данные, если нет, то добавляет данные в базу
// @Accept  json
// @Produce  json
// @Param body body entity.System true "Объект системы"
// @Success 200 {object} entity.System
// @Failure 404 {object} structure.GrpcError
// @Failure 409 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /system/create_update_system [POST]
func (systemController) CreateUpdateSystem(system entity.System) (*entity.System, error) {
	existed, err := model.SystemRep.GetSystemByName(system.Name)
	if err != nil {
		return nil, err
	}
	if system.Id == 0 {
		if existed != nil {
			return nil, status.Errorf(codes.AlreadyExists, "System with name %s already exists", system.Name)
		}

		system, err = model.SystemRep.CreateSystem(system)
		if err != nil {
			return nil, err
		}

		return &system, nil
	}

	if existed != nil && existed.Id != system.Id {
		return nil, status.Errorf(codes.AlreadyExists, "System with name %s already exists", system.Name)
	}

	existed, err = model.SystemRep.GetSystemById(system.Id)
	if err != nil {
		return nil, err
	}
	if existed == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("System with id %d not found", system.Id))
	}

	system, err = model.SystemRep.UpdateSystem(system)
	if err != nil {
		return nil, err
	}

	return &system, nil
}

// GetSystemById godoc
// @Tags system
// @Summary Получить систему по идентификатору
// @Description Возвращает описание системы по ее идентификатору
// @Accept  json
// @Produce  json
// @Param body body domain.Identity true "Идентификатор системы"
// @Success 200 {object} entity.System
// @Failure 404 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /system/get_system_by_id [POST]
func (systemController) GetSystemById(identity domain.Identity) (*entity.System, error) {
	sys, err := model.SystemRep.GetSystemById(identity.Id)
	if err != nil {
		return nil, err
	}
	if sys == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("System with id %d not found", identity.Id))
	}

	return sys, nil
}

// DeleteSystems godoc
// @Tags system
// @Summary Удалить системы
// @Description Удаляет системы по списку их идентификаторов, возвращает количество удаленных систем
// @Accept  json
// @Produce  json
// @Param body body []integer false "Массив идентификаторов систем"
// @Success 200 {object} domain.DeleteResponse
// @Failure 400 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /system/delete_systems [POST]
func (systemController) DeleteSystems(list []int32) (domain.DeleteResponse, error) {
	if len(list) == 0 {
		return domain.DeleteResponse{}, status.Error(codes.InvalidArgument, "At least one id are required")
	}

	res, err := model.SystemRep.DeleteSystems(list)
	if err != nil {
		return domain.DeleteResponse{}, err
	}

	return domain.DeleteResponse{Deleted: res}, nil
}
