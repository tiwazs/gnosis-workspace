package services

import (
	"crypto/rand"
	"time"

	"github.com/google/uuid"
	"github.com/tiwazs/gnosis-workspace/app/models"
	"gorm.io/gorm"
)

type WorkspaceService struct {
	db *gorm.DB
}

func NewWorkspaceService(db *gorm.DB) *WorkspaceService {
	return &WorkspaceService{db: db}
}

func (service *WorkspaceService) CreateWorkspace(name string, ownerID string) (*models.Workspace, error) {
	workspace := &models.Workspace{
		ID:      uuid.New(),
		Name:    name,
		OwnerID: ownerID,
	}

	var role struct {
		ID uuid.UUID
	}

	if err := service.db.Table("roles").Where("name = ?", "Owner").Select("id").First(&role).Error; err != nil {
		return nil, err
	}

	if err := service.db.Create(workspace).Error; err != nil {
		return nil, err
	}

	if err := service.db.Create(&models.WorkspaceMember{
		ID:          uuid.New(),
		WorkspaceID: workspace.ID,
		UserID:      ownerID,
    	RoleID:      role.ID,
	}).Error; err != nil {
		return nil, err
	}

	return workspace, nil
}

func (service *WorkspaceService) GetWorkspace(workspaceID uuid.UUID) (*models.Workspace, error) {
	var workspace models.Workspace
	if err := service.db.First(&workspace, "id = ?", workspaceID).Error; err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (service *WorkspaceService) GetWorkspacesByUserId(ownerID string) ([]models.Workspace, error) {
	var workspaces []models.Workspace
	if err := service.db.Where("owner_id = ?", ownerID).Find(&workspaces).Error; err != nil {
		return nil, err
	}

	return workspaces, nil
}

func (service *WorkspaceService) GenerateRegistrationToken(workspaceID, userID string) (*models.RegistrationToken, error) {
	parsedWorkspaceID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, err
	}

	token := generateCode()
	record := &models.RegistrationToken{
		Token:       token,
		WorkspaceID: parsedWorkspaceID,
		CreatedBy:   userID,
		ExpiresAt:   time.Now().UTC().Add(30 * time.Minute),
		Used:        false,
	}

	if err := service.db.Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func generateCode() string {
	alphabet := []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	result := make([]rune, 8)
	for index, randomByte := range randomBytes {
		result[index] = alphabet[int(randomByte)%len(alphabet)]
	}
	return "RPi-" + string(result[:4]) + "-" + string(result[4:])
}
