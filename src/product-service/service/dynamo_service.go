package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	// DynamoDB limits
	MaxBatchWriteItems = 25
	MaxRetryAttempts   = 3

	// Table creation timeout
	TableCreationTimeout = 5 * time.Minute
)

// PaginationToken can be used to abstract the LastEvaluatedKey for API responses
type PaginationToken struct {
	LastEvaluatedKey map[string]types.AttributeValue `json:"lastEvaluatedKey"`
}

// DynamoService provides a generic interface for DynamoDB operations
type DynamoService[T any] struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoService creates a new DynamoDB service instance
func NewDynamoService[T any](client *dynamodb.Client,
	tableName string) *DynamoService[T] {

	return &DynamoService[T]{
		client:    client,
		tableName: tableName,
	}
}

// TableDefinition holds table schema configuration
type TableDefinition struct {
	AttributeDefinitions   []types.AttributeDefinition
	KeySchema              []types.KeySchemaElement
	GlobalSecondaryIndexes []types.GlobalSecondaryIndex
	LocalSecondaryIndexes  []types.LocalSecondaryIndex
	BillingMode            types.BillingMode
	ProvisionedThroughput  *types.ProvisionedThroughput
}

// CreateTableWithDefinition creates a table with custom schema
func (s *DynamoService[T]) CreateTableWithDefinition(ctx context.Context, def TableDefinition) error {
	input := &dynamodb.CreateTableInput{
		TableName:            aws.String(s.tableName),
		AttributeDefinitions: def.AttributeDefinitions,
		KeySchema:            def.KeySchema,
		BillingMode:          def.BillingMode,
	}

	// Add GSI if provided
	if len(def.GlobalSecondaryIndexes) > 0 {
		input.GlobalSecondaryIndexes = def.GlobalSecondaryIndexes
	}

	// Add LSI if provided
	if len(def.LocalSecondaryIndexes) > 0 {
		input.LocalSecondaryIndexes = def.LocalSecondaryIndexes
	}

	// Add provisioned throughput if not using pay-per-request
	if def.BillingMode == types.BillingModeProvisioned && def.ProvisionedThroughput != nil {
		input.ProvisionedThroughput = def.ProvisionedThroughput
	}

	_, err := s.client.CreateTable(ctx, input)
	if err != nil {
		var resourceInUseEx *types.ResourceInUseException
		if errors.As(err, &resourceInUseEx) {
			fmt.Printf("Table already exists: %s\n", s.tableName)
			return nil
		}
		return fmt.Errorf("failed to create table %s: %w", s.tableName, err)
	}

	// Wait for table to be active
	return s.waitForTableActive(ctx)
}

// CreateTable creates a simple table with string ID as primary key (backward compatibility)
func (s *DynamoService[T]) CreateTable(ctx context.Context) error {
	def := TableDefinition{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	}

	return s.CreateTableWithDefinition(ctx, def)
}

func (s *DynamoService[T]) waitForTableActive(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, TableCreationTimeout)
	defer cancel()

	waiter := dynamodb.NewTableExistsWaiter(s.client)
	err := waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	}, TableCreationTimeout)

	if err != nil {
		return fmt.Errorf("failed waiting for table %s to be active: %w", s.tableName, err)
	}

	fmt.Printf("Table created successfully %s", s.tableName)
	return nil
}

// DeleteTable deletes the DynamoDB table
func (s *DynamoService[T]) DeleteTable(ctx context.Context) error {
	_, err := s.client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete table %s: %w", s.tableName, err)
	}
	return nil
}

// TableExists checks if the table exists
func (s *DynamoService[T]) TableExists(ctx context.Context) (bool, error) {
	_, err := s.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	})

	if err != nil {
		var notFoundEx *types.ResourceNotFoundException
		if errors.As(err, &notFoundEx) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check table existence for %s: %w", s.tableName, err)
	}

	return true, nil
}

// AddBatchItems adds multiple items in batches (handles DynamoDB 25-item limit)
func (s *DynamoService[T]) AddBatchItems(ctx context.Context, items []T) error {
	if len(items) == 0 {
		return nil
	}

	// Process items in batches of 25 (DynamoDB limit)
	for i := 0; i < len(items); i += MaxBatchWriteItems {
		end := i + MaxBatchWriteItems
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]
		if err := s.processBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to process batch %d-%d: %w", i, end-1, err)
		}
	}

	return nil
}

func (s *DynamoService[T]) processBatch(ctx context.Context, items []T) error {
	writeRequests := make([]types.WriteRequest, 0, len(items))

	for _, data := range items {
		item, err := attributevalue.MarshalMap(data)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %w", err)
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{Item: item},
		})
	}

	// Handle unprocessed items with exponential backoff
	unprocessedItems := map[string][]types.WriteRequest{
		s.tableName: writeRequests,
	}

	for attempt := 0; attempt < MaxRetryAttempts && len(unprocessedItems) > 0; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		result, err := s.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: unprocessedItems,
		})

		if err != nil {
			return fmt.Errorf("batch write failed on attempt %d: %w", attempt+1, err)
		}

		unprocessedItems = result.UnprocessedItems
	}

	if len(unprocessedItems) > 0 {
		return fmt.Errorf("failed to process all items after %d attempts, %d items remain unprocessed",
			MaxRetryAttempts, len(unprocessedItems[s.tableName]))
	}

	return nil
}

// GetItemConsistent retrieves a single item with strong consistency
func (s *DynamoService[T]) GetItemConsistent(ctx context.Context, key map[string]types.AttributeValue) (*T, error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(s.tableName),
		Key:            key,
		ConsistentRead: aws.Bool(true),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get item from table %s: %w", s.tableName, err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var item T
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return &item, nil
}

// DeleteItem removes an item from the table
func (s *DynamoService[T]) DeleteItem(ctx context.Context, key map[string]types.AttributeValue) error {
	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key:       key,
	})

	if err != nil {
		return fmt.Errorf("failed to delete item from table %s: %w", s.tableName, err)
	}

	return nil
}

// UpdateItemOptions provides configuration for update operations
type UpdateItemOptions struct {
	Key                  map[string]types.AttributeValue
	UpdateExpression     string
	ConditionExpression  *string
	ExpressionAttributes map[string]any
	ReturnValues         types.ReturnValue
}

// UpdateItem updates an item with comprehensive options
func (s *DynamoService[T]) UpdateItem(ctx context.Context, opts UpdateItemOptions) (*T, error) {
	update := expression.UpdateBuilder{}

	for key, value := range opts.ExpressionAttributes {
		switch v := value.(type) {
		case int, int64:
			update = update.Set(expression.Name(key), expression.Value(v))
		case float64:
			update = update.Set(expression.Name(key), expression.Value(v))
		default:
			update = update.Set(expression.Name(key), expression.Value(fmt.Sprintf("%v", v)))
		}
	}

	expr, err := expression.NewBuilder().WithUpdate(update).Build()

	if err != nil {
		return nil, fmt.Errorf("error when build update expression: %v", err)
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(s.tableName),
		Key:                       opts.Key,
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              opts.ReturnValues,
	}

	result, err := s.client.UpdateItem(ctx, input)
	if err != nil {
		var conditionalCheckEx *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckEx) {
			return nil, fmt.Errorf("update condition check failed: %w", err)
		}
		return nil, fmt.Errorf("failed to update item in table %s: %w", s.tableName, err)
	}

	// If caller wants the updated item back
	if opts.ReturnValues != types.ReturnValueNone && result.Attributes != nil {
		var updated T
		if err := attributevalue.UnmarshalMap(result.Attributes, &updated); err != nil {
			return nil, fmt.Errorf("failed to unmarshal updated item: %w", err)
		}
		return &updated, nil
	}

	return nil, nil
}

// Helper function to create a simple key for string IDs
func CreateStringKey(id string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: id},
	}
}

// Helper function to create a composite key
func CreateCompositeKey(partitionKey, sortKey string, partitionValue, sortValue interface{}) (map[string]types.AttributeValue, error) {
	key := make(map[string]types.AttributeValue)

	// Marshal partition key
	pkValue, err := attributevalue.Marshal(partitionValue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal partition key: %w", err)
	}
	key[partitionKey] = pkValue

	// Marshal sort key
	skValue, err := attributevalue.Marshal(sortValue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sort key: %w", err)
	}
	key[sortKey] = skValue

	return key, nil
}

// ScanItems performs a scan operation (use sparingly - prefer Query when possible)
func (s *DynamoService[T]) Scan(ctx context.Context, request ScanRequest) ([]T, error) {
	var items []T
	var response *dynamodb.ScanOutput

	expressionBuilder := expression.NewBuilder()

	if request.FilterBuilder != nil {
		expressionBuilder = expressionBuilder.WithFilter(*request.FilterBuilder)
	}

	if request.ProjectionBuilder != nil {
		expressionBuilder = expressionBuilder.WithProjection(*request.ProjectionBuilder)
	}

	expr, err := expressionBuilder.Build()

	if err != nil {
		fmt.Printf("Couldn't build expressions for scan. Here's why: %v\n", err)
	}

	scanPaginator := dynamodb.NewScanPaginator(s.client, &dynamodb.ScanInput{
		TableName:                 aws.String(s.tableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	})

	for scanPaginator.HasMorePages() {
		response, err = scanPaginator.NextPage(ctx)
		if err != nil {
			fmt.Printf("Couldn't scan for movies released between. Here's why: %v\n", err)
			break
		} else {
			var itemPage []T
			err = attributevalue.UnmarshalListOfMaps(response.Items, &itemPage)
			if err != nil {
				fmt.Printf("Couldn't unmarshal query response. Here's why: %v\n", err)
				break
			} else {
				items = append(items, itemPage...)
			}
		}
	}

	return items, err
}

type ScanRequest struct {
	FilterBuilder     *expression.ConditionBuilder
	ProjectionBuilder *expression.ProjectionBuilder
}

// Scan -> SELECT * FROM Products : scan all items before apply expression filter

// Query -> Filter base on partion key and sort key (optional) -> performance than Scan

// BatchWriteItem
func (s *DynamoService[T]) BatchWriteItems(ctx context.Context, items []T) (int, error) {
	var err error
	var item map[string]types.AttributeValue

	written := 0
	start := 0
	end := start + MaxBatchWriteItems

	for start < len(items) {
		var writeReqs []types.WriteRequest
		if end > len(items) {
			end = len(items)
		}

		for _, itemData := range items[start:end] {
			item, err = attributevalue.MarshalMap(itemData)
			if err != nil {
				fmt.Printf("Couldn't marshal item for batch writing. Here's why: %v\n", err)
			} else {
				writeReqs = append(
					writeReqs,
					types.WriteRequest{PutRequest: &types.PutRequest{Item: item}},
				)
			}
		}
		_, err = s.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{s.tableName: writeReqs}})

		if err != nil {
			log.Printf("Couldn't add a batch of items to %v. Here's why: %v\n", s.tableName, err)
		} else {
			written += len(writeReqs)
		}

		start = end
		end += MaxBatchWriteItems
	}

	return written, err
}

// GetItem retrieves a single item by key
func (s *DynamoService[T]) GetItem(ctx context.Context, key map[string]types.AttributeValue) (*T, error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(s.tableName),
		Key:            key,
		ConsistentRead: aws.Bool(false),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get item from table %s: %w", s.tableName, err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var item T
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return &item, nil
}

// PutItem adds a single item to the table
func (s *DynamoService[T]) PutItem(ctx context.Context, data T) error {
	item, err := attributevalue.MarshalMap(data)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to add item to table %s: %w", s.tableName, err)
	}

	return nil
}
