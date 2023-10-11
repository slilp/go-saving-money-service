package service

import (
	"github.com/jinzhu/copier"
	repository "github.com/slilp/go-saving-money-service/repository/planner"
)

type plannerService struct {
	plannerRepo repository.PlannerRepository
}

func NewPlannerService(plannerRepo repository.PlannerRepository) PlannerService {
	return plannerService{plannerRepo}
}

func (s plannerService) CreatePlanner(req CreateUpdatePlannerDto) (*PlannerInfo, error) {
	var srvRes PlannerInfo
	err := s.plannerRepo.Create(repository.PlannerEntity{
		Name:   req.Name,
		Target: req.Target,
	})
	if err != nil {
		return nil, err
	}

	copier.Copy(&srvRes, &req)
	return &srvRes, err
}
