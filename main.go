// @title           Gnosis Workspace API
// @version         1.0
// @description     Workspace service API
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description     Type "Bearer" followed by a space and the JWT from main-service login.
package main

import (
	"log"
	"fmt"
	"os"
	"net"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"google.golang.org/grpc"
	"github.com/tiwazs/gnosis-workspace/app/controllers"
	"github.com/tiwazs/gnosis-workspace/app/database"
	appgrpc "github.com/tiwazs/gnosis-workspace/app/grpc" // your RegistrationServer
	workspacev1 "github.com/tiwazs/gnosis-workspace/app/grpc/workspace/v1"
	"github.com/tiwazs/gnosis-workspace/app/services"
	_ "github.com/tiwazs/gnosis-workspace/docs"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db := database.Connect()
	database.RunMigrations()

	router := gin.Default()
	controllers.RegisterRoutes(router, db)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	workspaceService := services.NewWorkspaceService(db)

	// Start gRPC server
	listening_address, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("GRPC_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	workspacev1.RegisterRegistrationServiceServer(grpcServer, &appgrpc.RegistrationServer{
		Service: workspaceService,
	})
	go grpcServer.Serve(listening_address)

	router.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
}
