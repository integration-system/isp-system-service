package controller

import (
	"isp-system-service/domain"
	"isp-system-service/entity"
	"isp-system-service/model"

	_ "github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var Domain domainController

type domainController struct{}

// GetDomainsBySystemId godoc
// @Tags domain
// @Summary Получить домены по идентификатору системы
// @Description Возвращает список доменов по системному идентификатору
// @Accept  json
// @Produce  json
// @Param body body integer false "Идентификатор системы"
// @Success 200 {array} entity.Domain
// @Failure 500 {object} structure.GrpcError
// @Router /domain/get_domains_by_system_id [POST]
func (domainController) GetDomainsBySystemId(md metadata.MD) ([]entity.Domain, error) {
	sysId, err := utils.ResolveMetadataIdentity(utils.SystemIdHeader, md)
	if err != nil {
		return nil, err
	}

	return model.DomainRep.GetDomainsBySystemId(int32(sysId))
}

// CreateUpdateDomain godoc
// @Tags domain
// @Summary Создать/обновить домен
// @Description Если домен с такими идентификатором существует, то обновляет данные, если нет, то добавляет данные в базу
// @Accept  json
// @Produce  json
// @Param body body entity.Domain true "Объект домена"
// @Success 200 {object} entity.Domain
// @Failure 500 {object} structure.GrpcError
// @Router /domain/create_update_domain [POST]
func (domainController) CreateUpdateDomain(domain entity.Domain, md metadata.MD) (*entity.Domain, error) {
	existed, err := model.DomainRep.GetDomainByNameAndSystemId(domain.Name, domain.SystemId)
	if err != nil {
		return nil, err
	}

	sysId, err := utils.ResolveMetadataIdentity(utils.SystemIdHeader, md)
	if err != nil {
		return nil, err
	}
	domain.SystemId = int32(sysId)

	sys, err := model.SystemRep.GetSystemById(domain.SystemId)
	if err != nil {
		return nil, err
	}
	if sys == nil {
		return nil, status.Errorf(codes.InvalidArgument, "System with id %d not found", domain.SystemId)
	}

	if domain.Id == 0 {
		if existed != nil {
			return nil, status.Errorf(codes.AlreadyExists, "Domain with name %s already exists", domain.Name)
		}

		domain, err = model.DomainRep.CreateDomain(domain)
		if err != nil {
			return nil, err
		}

		return &domain, nil
	}

	if existed != nil && existed.Id != domain.Id {
		return nil, status.Errorf(codes.AlreadyExists, "Domain with name %s already exists", domain.Name)
	}

	existed, err = model.DomainRep.GetDomainById(domain.Id)
	if err != nil {
		return nil, err
	}
	if existed == nil {
		return nil, status.Errorf(codes.NotFound, "Domain with id %d not found", domain.Id)
	}

	domain, err = model.DomainRep.UpdateDomain(domain)
	if err != nil {
		return nil, err
	}

	return &domain, nil
}

// GetDomainById godoc
// @Tags domain
// @Summary Получить домен по идентификатору
// @Description Возвращает описание домена по его идентификатору
// @Accept  json
// @Produce  json
// @Param body body domain.Identity true "Идентификатор домена"
// @Success 200 {object} entity.Domain
// @Failure 404 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /domain/get_domain_by_id [POST]
func (domainController) GetDomainById(identity domain.Identity) (*entity.Domain, error) {
	d, err := model.DomainRep.GetDomainById(identity.Id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, status.Errorf(codes.NotFound, "Domain with id %d not found", identity.Id)
	}

	return d, err
}

// DeleteDomains godoc
// @Tags domain
// @Summary Удаление доменов
// @Description Удаляет домены по списку их идентификаторов, возвращает количество удаленных доменов
// @Accept  json
// @Produce  json
// @Param body body []integer false "Массив идентификаторов доменов"
// @Success 200 {object} domain.DeleteResponse
// @Failure 400 {object} structure.GrpcError
// @Failure 500 {object} structure.GrpcError
// @Router /domain/delete_domains [POST]
func (domainController) DeleteDomains(list []int32) (domain.DeleteResponse, error) {
	if len(list) == 0 {
		return domain.DeleteResponse{Deleted: 0}, status.Errorf(codes.InvalidArgument, "At least one id are required")
	}

	res, err := model.DomainRep.DeleteDomains(list)
	if err != nil {
		return domain.DeleteResponse{}, err
	}

	return domain.DeleteResponse{Deleted: res}, nil
}
