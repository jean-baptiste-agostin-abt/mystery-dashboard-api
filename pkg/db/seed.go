package db

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/jibe0123/mysteryfactory/internal/config"
	"github.com/jibe0123/mysteryfactory/internal/models"
)

// Seed inserts initial data if it does not already exist.
func Seed(gdb *gorm.DB, cfg *config.Config) error {
	tenantID := cfg.DefaultTenantID
	if tenantID == "" {
		tenantID = "default"
	}

	var count int64
	if err := gdb.Model(&models.Tenant{}).Where("id = ?", tenantID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		t := &models.Tenant{
			ID:        tenantID,
			Name:      "Default Tenant",
			Domain:    "default",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := gdb.Create(t).Error; err != nil {
			return err
		}
	}

	superEmail := "admin@example.com"
	if err := gdb.Model(&models.User{}).Where("email = ?", superEmail).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		u := &models.User{
			ID:        uuid.New().String(),
			TenantID:  tenantID,
			Email:     superEmail,
			Password:  string(hashed),
			FirstName: "Super",
			LastName:  "Admin",
			Role:      string(models.RoleAdmin),
			Status:    string(models.StatusActive),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := gdb.Create(u).Error; err != nil {
			return err
		}
		ws := &models.Workspace{
			ID:        uuid.New().String(),
			TenantID:  tenantID,
			UserID:    u.ID,
			Name:      "Default Workspace",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := gdb.Create(ws).Error; err != nil {
			return err
		}
	}
	return nil
}
