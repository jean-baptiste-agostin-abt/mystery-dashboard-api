package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jibe0123/mysteryfactory/internal/models"
)

// tenantRepository implements models.TenantRepository.
type tenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository creates a new repository.
func NewTenantRepository(db *gorm.DB) models.TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(t *models.Tenant) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return r.db.Create(t).Error
}

func (r *tenantRepository) GetByID(id string) (*models.Tenant, error) {
	var t models.Tenant
	err := r.db.First(&t, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, models.ErrTenantNotFound
	}
	return &t, err
}

func (r *tenantRepository) GetByDomain(domain string) (*models.Tenant, error) {
	var t models.Tenant
	err := r.db.First(&t, "domain = ?", domain).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, models.ErrTenantNotFound
	}
	return &t, err
}

func (r *tenantRepository) Update(t *models.Tenant) error {
	return r.db.Save(t).Error
}

func (r *tenantRepository) Delete(id string) error {
	return r.db.Delete(&models.Tenant{}, "id = ?", id).Error
}

func (r *tenantRepository) List(limit, offset int) ([]*models.Tenant, error) {
	var tenants []*models.Tenant
	err := r.db.Limit(limit).Offset(offset).Find(&tenants).Error
	return tenants, err
}
