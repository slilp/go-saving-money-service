package handler

import (
	"github.com/slilp/go-auth/middleware"
	repository "github.com/slilp/go-saving-money-service/repository/planner"
	service "github.com/slilp/go-saving-money-service/service/planner"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PlannerInitialize(
	db *gorm.DB,
	group *gin.RouterGroup,
	middlewareAuth middleware.AuthMiddleware) {
	repo := repository.NewAccountRepository(db)
	service := service.NewPlannerService(
		repo,
	)
	plannerServer(group, service, middlewareAuth)
}
