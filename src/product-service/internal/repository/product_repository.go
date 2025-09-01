package repository

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/quochao170402/ecommerce-aws/internal/domain"
	"github.com/quochao170402/ecommerce-aws/service"
)

type ProductRepository interface {
	BaseRepository[domain.Product]

	FindByCategory(ctx context.Context, categoryId string) ([]domain.Product, error)
	FindByBrand(ctx context.Context, brandId string) ([]domain.Product, error)
	SearchByName(ctx context.Context, keyword string) ([]domain.Product, error)
}

type productRepository struct {
	BaseRepository[domain.Product]
	dynamo *service.DynamoService[domain.Product]
}

func NewProductRepository(client *dynamodb.Client) ProductRepository {
	const tableName string = "Products"
	dynamoService := service.NewDynamoService[domain.Product](client, tableName)

	exist, err := dynamoService.TableExists(context.Background())
	if err != nil {
		log.Fatalf("Error when process TableExists: %v", err)
	}

	if !exist {
		if err := dynamoService.CreateTable(context.Background()); err != nil {
			log.Fatalf("Error when creating Products table: %v", err)
		}
	}

	return &productRepository{
		BaseRepository: NewBaseRepository[domain.Product](client, tableName),
		dynamo:         dynamoService,
	}
}

// SearchByName implements ProductRepository.
func (p *productRepository) SearchByName(ctx context.Context, keyword string) ([]domain.Product, error) {
	filtEx := expression.Contains(expression.Name("name"), keyword)
	projection := expression.NamesList(
		expression.Name("id"),
		expression.Name("name"),
		expression.Name("status"),
		expression.Name("categoryId"),
		expression.Name("brandId"),
	)

	request := service.ScanRequest{
		FilterBuilder:     &filtEx,
		ProjectionBuilder: &projection,
	}
	return p.dynamo.Scan(ctx, request)
}

// FindByBrand implements ProductRepository.
func (p *productRepository) FindByBrand(ctx context.Context, brandId string) ([]domain.Product, error) {
	filtEx := expression.Equal(expression.Name("brandId"), expression.Value(brandId))
	projection := expression.NamesList(
		expression.Name("id"),
		expression.Name("name"),
		expression.Name("status"),
		expression.Name("categoryId"),
		expression.Name("brandId"),
	)

	request := service.ScanRequest{
		FilterBuilder:     &filtEx,
		ProjectionBuilder: &projection,
	}
	return p.dynamo.Scan(ctx, request)
}

// FindByCategory implements ProductRepository.
func (p *productRepository) FindByCategory(ctx context.Context, categoryId string) ([]domain.Product, error) {
	filtEx := expression.Equal(expression.Name("categoryId"), expression.Value(categoryId))
	projection := expression.NamesList(
		expression.Name("id"),
		expression.Name("name"),
		expression.Name("status"),
		expression.Name("categoryId"),
		expression.Name("brandId"),
	)

	request := service.ScanRequest{
		FilterBuilder:     &filtEx,
		ProjectionBuilder: &projection,
	}
	return p.dynamo.Scan(ctx, request)
}
