package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jibe0123/mysteryfactory/internal/models"
)

type workspaceRepository struct {
	db *gorm.DB
}

// NewWorkspaceRepository creates a workspace repository.
func NewWorkspaceRepository(db *gorm.DB) models.WorkspaceRepository {
	return &workspaceRepository{db: db}
}

func (r *workspaceRepository) Create(w *models.Workspace) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return r.db.Create(w).Error
}

func (r *workspaceRepository) GetByID(tenantID, id string) (*models.Workspace, error) {
	var w models.Workspace
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&w).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, models.ErrNotFound
	}
	return &w, err
}

func (r *workspaceRepository) ListByUser(tenantID, userID string) ([]*models.Workspace, error) {
	var ws []*models.Workspace
	err := r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).Find(&ws).Error
	return ws, err
}

func (r *workspaceRepository) Update(w *models.Workspace) error {
	return r.db.Save(w).Error
}

func (r *workspaceRepository) Delete(tenantID, id string) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.Workspace{}).Error
}
