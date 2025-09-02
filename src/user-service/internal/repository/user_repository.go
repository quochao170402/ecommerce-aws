package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/quochao170402/ecommerce-aws/user-service/internal/models"
	"gorm.io/gorm"
)

type IUserRepository interface {
	IBaseRepository[models.User]
	GetByRole(ctx context.Context, roleID uuid.UUID) ([]models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByName(ctx context.Context, name string) ([]models.User, error)
}

type UserRepository struct {
	IBaseRepository[models.User] // generic CRUD
	db                           *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{
		IBaseRepository: NewBaseRepository[models.User](db),
		db:              db,
	}
}

func (r *UserRepository) GetByRole(ctx context.Context, roleID uuid.UUID) ([]models.User, error) {
	var users []models.User

	if err := r.db.WithContext(ctx).
		Where("role_id = ?", roleID).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByName(ctx context.Context, name string) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Where("name ILIKE ?", "%"+name+"%").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
