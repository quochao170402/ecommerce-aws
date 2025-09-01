package domain

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoEntity interface for entities that can be stored in DynamoDB
type DynamoEntity interface {
	GetKey() map[string]types.AttributeValue
	GetTableName() string
}

// TimestampedEntity interface for entities with timestamps
type TimestampedEntity interface {
	SetCreatedAt(timestamp int64)
	SetUpdatedAt(timestamp int64)
	GetCreatedAt() int64
	GetUpdatedAt() int64
}

// VersionedEntity interface for entities with optimistic locking
type VersionedEntity interface {
	GetVersion() int
	SetVersion(version int)
}
