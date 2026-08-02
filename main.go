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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tiwazs/gnosis-workspace/app/controllers"
	"github.com/tiwazs/gnosis-workspace/app/database"
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

	router.Run(":8000")
}
