package handler

import (
	"fmt"

	"github.com/slilp/go-auth/middleware"
	"github.com/slilp/go-saving-money-service/common/kafka"
	repository "github.com/slilp/go-saving-money-service/repository/planner"
	service "github.com/slilp/go-saving-money-service/service/planner"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PlannerInitialize(
	db *gorm.DB,
	group *gin.RouterGroup,
	middlewareAuth middleware.AuthMiddleware,
	pubsub kafka.KafkaPubSub) {
	repo := repository.NewAccountRepository(db)
	service := service.NewPlannerService(
		repo,
	)

	pubsub.Subscribe("blinkEvent", func(payload []byte) {
		fmt.Println(string(payload))
	})

	plannerServer(group, service, middlewareAuth)
}
