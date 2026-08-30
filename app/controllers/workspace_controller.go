package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tiwazs/gnosis-workspace/app/services"
	"gorm.io/gorm"
)

type WorkspaceController struct {
	service *services.WorkspaceService
}

type CreateWorkspaceRequest struct {
	Name string `json:"name" binding:"required" example:"My Lab"`
}

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	workspaceService := services.NewWorkspaceService(db)
	workspaceController := &WorkspaceController{service: workspaceService}

	api := router.Group("/workspace")
	//api.Use(middleware.Auth())
	{
		api.POST("/workspaces", workspaceController.CreateWorkspace)
		api.GET("/workspaces", workspaceController.GetWorkspaces)
		//api.GET("/workspaces/:id", workspaceController.GetWorkspace)
		api.POST("/workspaces/:id/devices/token", workspaceController.GenerateToken)
	}
}

// CreateWorkspace godoc
// @Summary      Create a workspace
// @Description  Create a new workspace for the current user
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        body  body  CreateWorkspaceRequest  true  "Workspace"
// @Success      201   {object}  models.Workspace
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /workspaces [post]
func (controller *WorkspaceController) CreateWorkspace(context *gin.Context) {
	var body CreateWorkspaceRequest


	userID := context.GetHeader("X-User-ID")
	if err := context.ShouldBindJSON(&body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := controller.service.CreateWorkspace(body.Name, userID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "could not create workspace"})
		return
	}

	context.JSON(http.StatusCreated, workspace)
}

// GetWorkspaces godoc
// @Summary      List current user's workspaces
// @Description  Returns workspaces owned by the authenticated user
// @Tags         workspaces
// @Produce      json
// @Success      200  {array}   models.Workspace
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /workspaces [get]
func (controller *WorkspaceController) GetWorkspaces(context *gin.Context) {
	userID := context.GetHeader("X-User-ID")

	workspaces, err := controller.service.GetWorkspacesByUserId(userID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "could not list workspaces"})
		return
	}

	context.JSON(http.StatusOK, workspaces)
}

// GenerateToken godoc
// @Summary      Generate device registration token
// @Description  Create a short-lived registration token for provisioning a device
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Workspace ID" Format(uuid)
// @Success      201  {object}  models.RegistrationToken
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /workspaces/{id}/devices/token [post]
func (controller *WorkspaceController) GenerateToken(context *gin.Context) {
	workspaceID := context.Param("id")
	userID := context.GetString("user_id")

	token, err := controller.service.GenerateRegistrationToken(workspaceID, userID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	context.JSON(http.StatusCreated, token)
}
