package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jibe0123/mysteryfactory/internal/models"
)

// userRepository implements models.UserRepository using GORM.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *gorm.DB) models.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(tenantID, id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, models.ErrUserNotFound
	}
	return &user, err
}

func (r *userRepository) GetByEmail(tenantID, email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("tenant_id = ? AND email = ?", tenantID, email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, models.ErrUserNotFound
	}
	return &user, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(tenantID, id string) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.User{}).Error
}

func (r *userRepository) List(tenantID string, limit, offset int) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateLastLogin(tenantID, id string) error {
	return r.db.Model(&models.User{}).Where("tenant_id = ? AND id = ?", tenantID, id).Update("last_login", gorm.Expr("NOW()")).Error
}
