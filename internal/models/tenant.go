package models

import (
	"database/sql"
	"time"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID        string       `json:"id" db:"id"`
	Name      string       `json:"name" db:"name"`
	Domain    string       `json:"domain" db:"domain"`
	Settings  string       `json:"settings" db:"settings"` // JSON string
	Status    string       `json:"status" db:"status"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TenantRepository defines the interface for tenant operations
type TenantRepository interface {
	Create(tenant *Tenant) error
	GetByID(id string) (*Tenant, error)
	GetByDomain(domain string) (*Tenant, error)
	Update(tenant *Tenant) error
	Delete(id string) error
	List(limit, offset int) ([]*Tenant, error)
}

// TenantService handles business logic for tenants
type TenantService struct {
	repo TenantRepository
}

// NewTenantService creates a new tenant service
func NewTenantService(repo TenantRepository) *TenantService {
	return &TenantService{repo: repo}
}

// CreateTenant creates a new tenant
func (s *TenantService) CreateTenant(tenant *Tenant) error {
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	tenant.Status = "active"
	return s.repo.Create(tenant)
}

// GetTenant retrieves a tenant by ID
func (s *TenantService) GetTenant(id string) (*Tenant, error) {
	return s.repo.GetByID(id)
}

// GetTenantByDomain retrieves a tenant by domain
func (s *TenantService) GetTenantByDomain(domain string) (*Tenant, error) {
	return s.repo.GetByDomain(domain)
}

// UpdateTenant updates an existing tenant
func (s *TenantService) UpdateTenant(tenant *Tenant) error {
	tenant.UpdatedAt = time.Now()
	return s.repo.Update(tenant)
}

// DeleteTenant soft deletes a tenant
func (s *TenantService) DeleteTenant(id string) error {
	return s.repo.Delete(id)
}

// ListTenants retrieves a list of tenants with pagination
func (s *TenantService) ListTenants(limit, offset int) ([]*Tenant, error) {
	return s.repo.List(limit, offset)
}
