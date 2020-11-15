package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
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

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var form PutEntryForm
	err := json.Unmarshal([]byte(request.Body), &form)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if len(form.Name) <= 0 || len(form.URL) <= 0 {
		return events.APIGatewayProxyResponse{}, errors.New("Invalid data")
	}

	sess, err := session.NewSession()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	svc := dynamodb.New(sess)

	item, err := dynamodbattribute.MarshalMap(&Entry{
		Name:      form.Name,
		URL:       form.URL,
		CreatedAt: int64(time.Now().Unix()),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	result, err := svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("hkt-sh-entries"),
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
	lambda.Start(handler)
}
