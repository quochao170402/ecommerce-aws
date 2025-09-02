package repository

import (
	"context"

	"github.com/quochao170402/ecommerce-aws/user-service/internal/models"
	"gorm.io/gorm"
)

type IRefreshTokenRepository interface {
	IBaseRepository[models.RefreshToken]
	GetByToken(ctx context.Context, token string) (*models.RefreshToken, error)
}

// RefreshTokenRepository implements IRefreshTokenRepository
type RefreshTokenRepository struct {
	IBaseRepository[models.RefreshToken]
	db *gorm.DB
}

// constructor
func NewRefreshTokenRepository(db *gorm.DB) IRefreshTokenRepository {
	return &RefreshTokenRepository{
		IBaseRepository: NewBaseRepository[models.RefreshToken](db),
		db:              db,
	}
}

func (r *RefreshTokenRepository) GetByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	if err := r.db.WithContext(ctx).
		Where("token = ?", token).
		First(&refreshToken).Error; err != nil {
		return nil, err
	}
	return &refreshToken, nil
}
