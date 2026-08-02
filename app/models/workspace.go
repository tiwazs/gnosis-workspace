package models

import (
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	// OwnerID is the main-service user id (cuid), not a UUID.
	OwnerID   string    `gorm:"type:text;not null" json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WorkspaceMember struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;not null" json:"workspace_id"`
	UserID      string    `gorm:"type:text;not null" json:"user_id"`
    RoleID      uuid.UUID `gorm:"type:uuid;not null" json:"role_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type RegistrationToken struct {
	Token        string     `gorm:"primaryKey" json:"token"`
	WorkspaceID  uuid.UUID  `gorm:"type:uuid;not null" json:"workspace_id"`
	CreatedBy    string     `gorm:"type:text;not null" json:"created_by"`
	ExpiresAt    time.Time  `gorm:"not null" json:"expires_at"`
	Used         bool       `gorm:"default:false" json:"used"`
	UsedAt       *time.Time `json:"used_at"`
	UsedByDevice *uuid.UUID `gorm:"type:uuid" json:"used_by_device"`
}
