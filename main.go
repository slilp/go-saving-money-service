package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slilp/go-auth/common/postgres"
	"github.com/slilp/go-auth/middleware"
	"github.com/slilp/go-saving-money-service/common/kafka"
	handler "github.com/slilp/go-saving-money-service/handler/planner"
	service "github.com/slilp/go-saving-money-service/service/planner"
	"github.com/spf13/viper"
)

func main() {
	initTimeZone()
	initConfig()
	router := initGin()
	pubsub := kafka.NewRelayer()
	defer pubsub.Stop()
	initApplication(router, pubsub)
	server := runServer(router)
	shutdownServer(server)
}

func initApplication(router *gin.Engine, pubsub kafka.KafkaPubSub) {

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})

	db, err := postgres.NewPostgresDB()
	if err != nil {
		panic(fmt.Errorf("failed to create database connection: %w", err))
	}
	authMiddleware := middleware.NewJwtMiddleware(viper.GetString("jwt.accessToken"), viper.GetString("jwt.refreshToken"))
	handler.PlannerInitialize(db, &router.RouterGroup, authMiddleware, pubsub)
	pubsub.Start()
	testStr := service.CreateUpdatePlannerDto{
		Name:   "slil",
		Target: 111,
	}
	pubsub.Publish(context.Background(), "blinkEvent", testStr)
}

func initGin() *gin.Engine {

	if viper.GetString("app.mode") == "develop" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)

	}

	router := gin.New()
	router.Use(
		gin.Recovery(),
	)

	return router
}

func runServer(router *gin.Engine) *http.Server {

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("app.port")),
		Handler: router,
	}
	fmt.Printf("Server is running on port: %s\n", server.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatal("ListenAndServe")
		}
	}()

	return server
}

func shutdownServer(server *http.Server) {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	stop()
	// log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown")
	}
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

}

func initTimeZone() {
	ict, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatal(err)
	}

	time.Local = ict
}
