package domain

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

type ImageUrl struct {
	URL string `dynamodbav:"url" json:"url"`
	Alt string `dynamodbav:"alt" json:"alt"`
}

type Product struct {
	ID          string     `dynamodbav:"id" json:"id"`
	Name        string     `dynamodbav:"name" json:"name"`
	Price       float64    `dynamodbav:"price" json:"price"`
	Description string     `dynamodbav:"description" json:"description"`
	CreatedAt   int64      `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt   int64      `dynamodbav:"updatedAt" json:"updatedAt"`
	Status      string     `dynamodbav:"status" json:"status"`
	Version     int        `dynamodbav:"version" json:"version"`
	BrandID     string     `dynamodbav:"brandId" json:"brandId"`
	CategoryID  string     `dynamodbav:"categoryId" json:"categoryId"`
	Images      []ImageUrl `dynamodbav:"imageUrls" json:"imageUrls"`
}

// Implement DynamoEntity interface for Product
func (p Product) GetKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: p.ID},
	}
}

func (p Product) GetTableName() string {
	return "products"
}

// Implement TimestampedEntity interface for Product
func (p *Product) SetCreatedAt(timestamp int64) { p.CreatedAt = timestamp }
func (p *Product) SetUpdatedAt(timestamp int64) { p.UpdatedAt = timestamp }
func (p Product) GetCreatedAt() int64           { return p.CreatedAt }
func (p Product) GetUpdatedAt() int64           { return p.UpdatedAt }

// Implement VersionedEntity interface for Product
func (p Product) GetVersion() int         { return p.Version }
func (p *Product) SetVersion(version int) { p.Version = version }
func (p *Product) IncrementVersion()      { p.Version++ }
