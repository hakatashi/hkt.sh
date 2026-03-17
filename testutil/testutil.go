package testutil

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// NewDynamoDBClient creates a DynamoDB client pointing at DynamoDB Local.
func NewDynamoDBClient(endpoint string) *dynamodb.DynamoDB {
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(endpoint),
		Region:      aws.String("ap-northeast-1"),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
	}))
	return dynamodb.New(sess)
}

// CreateTable creates the given table in DynamoDB Local.
func CreateTable(db *dynamodb.DynamoDB, tableName string) error {
	_, err := db.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String("Name"), AttributeType: aws.String("S")},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("Name"), KeyType: aws.String("HASH")},
		},
		BillingMode: aws.String("PAY_PER_REQUEST"),
	})
	return err
}

// DeleteTable removes the given table from DynamoDB Local.
func DeleteTable(db *dynamodb.DynamoDB, tableName string) error {
	_, err := db.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	return err
}
