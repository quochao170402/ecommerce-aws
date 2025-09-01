package repository

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/quochao170402/ecommerce-aws/internal/domain"
	"github.com/quochao170402/ecommerce-aws/service"
)

// UpdateOptions provides configuration for update operations
type UpdateOptions struct {
	Key                  map[string]types.AttributeValue
	ConditionExpression  *string
	ExpressionAttributes map[string]interface{}
	ReturnValues         types.ReturnValue
}

// QueryOptions provides configuration for query operations
type QueryOptions struct {
	IndexName                 *string
	KeyConditionExpression    *string
	FilterExpression          *string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]types.AttributeValue
	ProjectionExpression      *string
	ScanIndexForward          *bool
	Limit                     *int32
	ConsistentRead            *bool
}

// PaginatedQueryOptions extends QueryOptions with pagination support
type PaginatedQueryOptions struct {
	QueryOptions
	ExclusiveStartKey map[string]types.AttributeValue
	PageSize          int32
}

// baseRepository implements BaseRepository interface
type baseRepository[T domain.DynamoEntity] struct {
	service *service.DynamoService[T]
}

// BaseRepository defines the common operations for all entities
type BaseRepository[T domain.DynamoEntity] interface {
	// Basic CRUD operations
	Save(ctx context.Context, entity *T) error
	SaveBatch(ctx context.Context, entities *[]T) error
	FindByID(ctx context.Context, id string) (*T, error)
	FindByIDConsistent(ctx context.Context, id string) (*T, error)
	Delete(ctx context.Context, entity T) error
	DeleteByID(ctx context.Context, id string) error

	// Advanced operations
	Update(ctx context.Context, entity *T, opts UpdateOptions) (*T, error)
	UpdateByID(ctx context.Context, id string, opts UpdateOptions) (*T, error)
	Query(ctx context.Context, opts QueryOptions) ([]T, error)
	QueryWithPaging(ctx context.Context, opts PaginatedQueryOptions) (*service.QueryResult[T], error)

	// Conditional operations
	SaveIfNotExists(ctx context.Context, entity T) error
	// UpdateWithOptimisticLock(ctx context.Context, entity T, expectedVersion int, updates map[string]interface{}) (*T, error)

	// Utility operations
	Exists(ctx context.Context, id string) (bool, error)
	Count(ctx context.Context, opts QueryOptions) (int32, error)

	ScanItems(ctx context.Context) ([]T, error)
}

// NewBaseRepository creates a new base repository instance
func NewBaseRepository[T domain.DynamoEntity](client *dynamodb.Client, tableName string) BaseRepository[T] {

	dynamoService := service.NewDynamoService[T](client, tableName)

	exist, err := dynamoService.TableExists(context.Background())

	if err != nil {
		log.Fatalf("Error when process TableExists: %v", err)
	}

	if !exist {
		dynamoService.CreateTable(context.Background())
	}

	return &baseRepository[T]{service: dynamoService}
}

// Save saves an entity with automatic timestamps
func (r *baseRepository[T]) Save(ctx context.Context, entity *T) error {
	// Set timestamps if the entity supports it
	if timestamped, ok := any(entity).(domain.TimestampedEntity); ok {
		now := time.Now().Unix()
		timestamped.SetCreatedAt(now)
		timestamped.SetUpdatedAt(now)
	}

	if versioned, ok := any(entity).(domain.VersionedEntity); ok {
		versioned.SetVersion(1)
	}

	return r.service.AddItem(ctx, *entity)
}

func (r *baseRepository[T]) SaveBatch(ctx context.Context, entities *[]T) error {
	items := *entities
	now := time.Now().Unix()

	for i := range items {
		// Take pointer to each item
		if timestamped, ok := any(&items[i]).(domain.TimestampedEntity); ok {
			timestamped.SetCreatedAt(now)
			timestamped.SetUpdatedAt(now)
		}

		if versioned, ok := any(&items[i]).(domain.VersionedEntity); ok {
			versioned.SetVersion(1)
		}
	}

	return r.service.AddBatchItems(ctx, items)
}

// FindByID finds an entity by its ID (eventually consistent)
func (r *baseRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	key := service.CreateStringKey(id)
	return r.service.GetItem(ctx, key)
}

// FindByIDConsistent finds an entity by its ID with strong consistency
func (r *baseRepository[T]) FindByIDConsistent(ctx context.Context, id string) (*T, error) {
	key := service.CreateStringKey(id)
	return r.service.GetItemConsistent(ctx, key)
}

// Delete removes an entity
func (r *baseRepository[T]) Delete(ctx context.Context, entity T) error {
	return r.service.DeleteItem(ctx, entity.GetKey())
}

// DeleteByID removes an entity by its ID
func (r *baseRepository[T]) DeleteByID(ctx context.Context, id string) error {
	key := service.CreateStringKey(id)
	return r.service.DeleteItem(ctx, key)
}

// Update updates an entity with custom options
func (r *baseRepository[T]) Update(ctx context.Context, entity *T, opts UpdateOptions) (*T, error) {
	attributes := opts.ExpressionAttributes
	// Set timestamps if the entity supports it
	if _, ok := any(entity).(domain.TimestampedEntity); ok {
		now := time.Now().Unix()
		attributes["updatedAt"] = now
	}

	if versioned, ok := any(entity).(domain.VersionedEntity); ok {
		attributes["version"] = versioned.GetVersion() + 1
	}

	return r.service.UpdateItem(ctx, service.UpdateItemOptions{
		Key:                  (*entity).GetKey(),
		ConditionExpression:  opts.ConditionExpression,
		ExpressionAttributes: attributes,
		ReturnValues:         opts.ReturnValues,
	})
}

// UpdateByID updates an entity by ID with custom options
func (r *baseRepository[T]) UpdateByID(ctx context.Context, id string, opts UpdateOptions) (*T, error) {
	key := service.CreateStringKey(id)
	return r.service.UpdateItem(ctx, service.UpdateItemOptions{
		Key:                  key,
		ConditionExpression:  opts.ConditionExpression,
		ExpressionAttributes: opts.ExpressionAttributes,
		ReturnValues:         opts.ReturnValues,
	})
}

// Query performs a query operation
func (r *baseRepository[T]) Query(ctx context.Context, opts QueryOptions) ([]T, error) {
	return r.service.QueryItems(ctx, service.QueryOptions{
		IndexName:                 opts.IndexName,
		KeyConditionExpression:    opts.KeyConditionExpression,
		FilterExpression:          opts.FilterExpression,
		ExpressionAttributeNames:  opts.ExpressionAttributeNames,
		ExpressionAttributeValues: opts.ExpressionAttributeValues,
		ProjectionExpression:      opts.ProjectionExpression,
		ScanIndexForward:          opts.ScanIndexForward,
		Limit:                     opts.Limit,
		ConsistentRead:            opts.ConsistentRead,
	})
}

// QueryWithPaging performs a paginated query
func (r *baseRepository[T]) QueryWithPaging(ctx context.Context, opts PaginatedQueryOptions) (*service.QueryResult[T], error) {
	queryRequest := service.QueryRequest{
		IndexName:                 opts.IndexName,
		KeyConditionExpression:    opts.KeyConditionExpression,
		FilterExpression:          opts.FilterExpression,
		ExpressionAttributeNames:  opts.ExpressionAttributeNames,
		ExpressionAttributeValues: opts.ExpressionAttributeValues,
		ProjectionExpression:      opts.ProjectionExpression,
		ScanIndexForward:          opts.ScanIndexForward,
		Limit:                     opts.PageSize,
		ExclusiveStartKey:         opts.ExclusiveStartKey,
	}

	return r.service.QueryWithPaging(ctx, queryRequest)
}

// SaveIfNotExists saves an entity only if it doesn't already exist
func (r *baseRepository[T]) SaveIfNotExists(ctx context.Context, entity T) error {
	// Set timestamps
	if timestamped, ok := any(entity).(domain.TimestampedEntity); ok {
		now := time.Now().Unix()
		timestamped.SetCreatedAt(now)
		timestamped.SetUpdatedAt(now)
	}

	return r.service.AddItemWithCondition(ctx, entity, "attribute_not_exists(id)", nil)
}

// UpdateWithOptimisticLock updates with version checking for optimistic locking
// func (r *baseRepository[T]) UpdateWithOptimisticLock(ctx context.Context, entity T, expectedVersion int, updates map[string]interface{}) (*T, error) {
// 	// Build update expression
// 	updateExpr := "SET "
// 	exprAttrValues := make(map[string]types.AttributeValue)
// 	exprAttrNames := make(map[string]string)

// 	first := true
// 	for field, value := range updates {
// 		if !first {
// 			updateExpr += ", "
// 		}

// 		placeholder := fmt.Sprintf(":val_%s", field)
// 		attrValue, err := attributevalue.Marshal(value)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to marshal update value for %s: %w", field, err)
// 		}

// 		updateExpr += fmt.Sprintf("#%s = %s", field, placeholder)
// 		exprAttrValues[placeholder] = attrValue
// 		exprAttrNames["#"+field] = field
// 		first = false
// 	}

// 	// Add updatedAt and increment version
// 	updateExpr += ", #updatedAt = :updatedAt, #version = #version + :inc"
// 	exprAttrValues[":updatedAt"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())}
// 	exprAttrValues[":inc"] = &types.AttributeValueMemberN{Value: "1"}
// 	exprAttrValues[":expectedVersion"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", expectedVersion)}
// 	exprAttrNames["#updatedAt"] = "updatedAt"
// 	exprAttrNames["#version"] = "version"

// 	return r.service.UpdateItem(ctx, service.UpdateItemOptions{
// 		Key:                  entity.GetKey(),
// 		UpdateExpression:     updateExpr,
// 		ConditionExpression:  aws.String("#version = :expectedVersion"),
// 		ExpressionAttributes: exprAttrNames,
// 		ReturnValues:         types.ReturnValueAllNew,
// 	})
// }

// Exists checks if an entity exists by ID
func (r *baseRepository[T]) Exists(ctx context.Context, id string) (bool, error) {
	key := service.CreateStringKey(id)
	result, err := r.service.GetItem(ctx, key)
	if err != nil {
		return false, err
	}
	return result != nil, nil
}

// Count returns the count of items matching the query
func (r *baseRepository[T]) Count(ctx context.Context, opts QueryOptions) (int32, error) {
	// Use Select COUNT for better performance
	countOpts := service.QueryOptions{
		IndexName:                 opts.IndexName,
		KeyConditionExpression:    opts.KeyConditionExpression,
		FilterExpression:          opts.FilterExpression,
		ExpressionAttributeNames:  opts.ExpressionAttributeNames,
		ExpressionAttributeValues: opts.ExpressionAttributeValues,
		// Project only the key to minimize data transfer
		ProjectionExpression: aws.String("id"),
	}

	items, err := r.service.QueryItems(ctx, countOpts)
	if err != nil {
		return 0, err
	}

	return int32(len(items)), nil
}

func (r *baseRepository[T]) ScanItems(ctx context.Context) ([]T, error) {
	return r.service.ScanItems(ctx, service.ScanOptions{})
}
