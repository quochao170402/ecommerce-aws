package repository

import (
	"context"

	"github.com/quochao170402/ecommerce-aws/user-service/internal/models"
	"gorm.io/gorm"
)

type IRoleRepository interface {
	IBaseRepository[models.Role]
	GetByName(ctx context.Context, name string) (*models.Role, error)
}

// RoleRepository implements IRoleRepository
type RoleRepository struct {
	IBaseRepository[models.Role]
	db *gorm.DB
}

// constructor
func NewRoleRepository(db *gorm.DB) IRoleRepository {
	return &RoleRepository{
		IBaseRepository: NewBaseRepository[models.Role](db),
		db:              db,
	}
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
