package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"

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
	rawName, ok := request.PathParameters["name"]
	if !ok {
		return events.APIGatewayProxyResponse{}, errors.New("Name parameter not found")
	}

	name, err := url.QueryUnescape(rawName)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	item, err := a.db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(a.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {S: aws.String(name)},
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if item.Item == nil {
		return events.APIGatewayProxyResponse{
			Body:       "<html><head><title>hkt.sh</title></head><body><h1>The entry you requested was not found.</h1></body></html>",
			StatusCode: 404,
			Headers: map[string]string{
				"Content-Type": "text/html; charset=utf-8",
			},
		}, nil
	}

	entry := Entry{}
	err = dynamodbattribute.UnmarshalMap(item.Item, &entry)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if entry.URL == "" {
		return events.APIGatewayProxyResponse{}, errors.New("Not found")
	}

	_, err = a.db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(a.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {S: aws.String(name)},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":inc": {N: aws.String("1")},
		},
		UpdateExpression: aws.String("ADD AccessCount :inc"),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("<html>\n<head><title>hkt.sh</title></head>\n<body><a href=\"%v\">moved here</a></body>\n</html>", entry.URL),
		StatusCode: 301,
		Headers: map[string]string{
			"Location":      entry.URL,
			"Cache-Control": "private, max-age=90",
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
