package domain

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

// Updated Brand struct
type Brand struct {
	Id        string `dynamodbav:"id" json:"id"`
	Name      string `dynamodbav:"name" json:"name"`
	CreatedAt int64  `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt int64  `dynamodbav:"updatedAt" json:"updatedAt"`
	Version   int    `dynamodbav:"version" json:"version"`
}

// Implement DynamoEntity interface for Brand
func (b Brand) GetKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: b.Id},
	}
}

func (b Brand) GetTableName() string {
	return "brands"
}

// Implement TimestampedEntity interface for Brand
func (b *Brand) SetCreatedAt(timestamp int64) { b.CreatedAt = timestamp }
func (b *Brand) SetUpdatedAt(timestamp int64) { b.UpdatedAt = timestamp }
func (b Brand) GetCreatedAt() int64           { return b.CreatedAt }
func (b Brand) GetUpdatedAt() int64           { return b.UpdatedAt }

// Implement VersionedEntity interface for Brand
func (b Brand) GetVersion() int         { return b.Version }
func (b *Brand) SetVersion(version int) { b.Version = version }
func (b *Brand) IncrementVersion()      { b.Version++ }
