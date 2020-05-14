package main

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var ()

type Entry struct {
	Name      string
	Url       string
	CreatedAt int64
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	rawName, ok := request.PathParameters["name"]
	if !ok {
		return events.APIGatewayProxyResponse{}, errors.New("Name parameter not found")
	}

	name, err := url.QueryUnescape(rawName)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	sess, err := session.NewSession()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	svc := dynamodb.New(sess)

	getParams := &dynamodb.GetItemInput{
		TableName: aws.String("hkt-sh-entries"),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String(name),
			},
		},
	}

	item, err := svc.GetItem(getParams)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	entry := Entry{}
	err = dynamodbattribute.UnmarshalMap(item.Item, &entry)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if entry.Url == "" {
		return events.APIGatewayProxyResponse{}, errors.New("Not found")
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("<html>\n<head><title>hkt.sh</title></head>\n<body><a href=\"%v\">moved here</a></body>\n</html>", entry.Url),
		StatusCode: 301,
		Headers: map[string]string{
			"Location":      entry.Url,
			"Cache-Control": "private, max-age=90",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
