package main

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Entry struct {
	Name      string
	URL       string
	CreatedAt int64
}

type PutEntryForm struct {
	Name string
	URL  string
}

type App struct {
	db        *dynamodb.DynamoDB
	tableName string
}

func newApp() (*App, error) {
	cfg := &aws.Config{}
	if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
		cfg.Endpoint = aws.String(endpoint)
		cfg.Credentials = credentials.NewStaticCredentials("test", "test", "")
		cfg.Region = aws.String("ap-northeast-1")
	}
	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}
	tableName := os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "hkt-sh-entries"
	}
	return &App{db: dynamodb.New(sess), tableName: tableName}, nil
}

func (a *App) handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var form PutEntryForm
	err := json.Unmarshal([]byte(request.Body), &form)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if len(form.Name) <= 0 || len(form.URL) <= 0 {
		return events.APIGatewayProxyResponse{}, errors.New("Invalid data")
	}

	item, err := dynamodbattribute.MarshalMap(&Entry{
		Name:      form.Name,
		URL:       form.URL,
		CreatedAt: int64(time.Now().Unix()),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	result, err := a.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(a.tableName),
		Item:      item,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	result_body, err := json.MarshalIndent(result, "", "\t")
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		Body:       string(result_body),
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	app, err := newApp()
	if err != nil {
		panic(err)
	}
	lambda.Start(app.handler)
}
