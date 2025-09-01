package domain

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

// Updated Category struct
type Category struct {
	Id        string `dynamodbav:"id" json:"id"`
	Name      string `dynamodbav:"name" json:"name"`
	CreatedAt int64  `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt int64  `dynamodbav:"updatedAt" json:"updatedAt"`
	Version   int    `dynamodbav:"version" json:"version"`
}

// Implement DynamoEntity interface for Category
func (c Category) GetKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: c.Id},
	}
}

func (c Category) GetTableName() string {
	return "categories"
}

// Implement TimestampedEntity interface for Category
func (c *Category) SetCreatedAt(timestamp int64) { c.CreatedAt = timestamp }
func (c *Category) SetUpdatedAt(timestamp int64) { c.UpdatedAt = timestamp }
func (c Category) GetCreatedAt() int64           { return c.CreatedAt }
func (c Category) GetUpdatedAt() int64           { return c.UpdatedAt }

// Implement VersionedEntity interface for Category
func (c Category) GetVersion() int         { return c.Version }
func (c *Category) SetVersion(version int) { c.Version = version }
func (c *Category) IncrementVersion()      { c.Version++ }
