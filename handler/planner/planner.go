package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slilp/go-auth/middleware"
	service "github.com/slilp/go-saving-money-service/service/planner"
	"github.com/slilp/go-saving-money-service/utils"
)

func plannerServer(router *gin.RouterGroup, s service.PlannerService, middlewareAuth middleware.AuthMiddleware) {
	handlerRoute := NewPlannerHttpHandler(s)
	authGroup := router.Group("/planner", middlewareAuth.Authentication())
	authGroup.POST("/new", handlerRoute.CreateNewPlanner)

}

func NewPlannerHttpHandler(service service.PlannerService) handler {
	return handler{service: service}
}

type handler struct {
	service service.PlannerService
}

func (h handler) CreateNewPlanner(ctx *gin.Context) {
	req := service.CreateUpdatePlannerDto{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BadRequest(err.Error()))
		return
	}

	srvRes, err := h.service.CreatePlanner(req)
	if err != nil {
		utils.ReturnError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, srvRes)
}
