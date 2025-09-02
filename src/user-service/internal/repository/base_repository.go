package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IBaseRepository[T any] interface {
	GetMany(ctx context.Context, filter map[string]interface{}) ([]T, error)
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type BaseRepository[T any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) IBaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) GetMany(ctx context.Context, filter map[string]any) ([]T, error) {
	var entities []T
	result := r.db.WithContext(ctx).Where(filter).Find(&entities)
	if result.Error != nil {
		return nil, result.Error
	}
	return entities, nil
}

func (r *BaseRepository[T]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	var entity T
	result := r.db.WithContext(ctx).Delete(&entity, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("entity not found")
	}
	return nil
}
