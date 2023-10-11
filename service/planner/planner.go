package service

type CreateUpdatePlannerDto struct {
	Name   string `json:"name" binding:"required"`
	Target uint   `json:"target" binding:"required"`
}

type PlannerInfo struct {
	Name   string `json:"name"`
	Target uint   `json:"target"`
}

type PlannerService interface {
	CreatePlanner(CreateUpdatePlannerDto) (*PlannerInfo, error)
}
