package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

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
	URL       string
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

	// Subdomain redirect: aaa.hkt.sh/bbb/ccc → hkt.sh/aaa/bbb/ccc
	host := request.Headers["Host"]
	var subdomainName string
	if strings.HasSuffix(host, ".hkt.sh") {
		subdomainName = strings.TrimSuffix(host, ".hkt.sh")
	} else if strings.HasSuffix(host, ".hkt.si") {
		subdomainName = strings.TrimSuffix(host, ".hkt.si")
	}
	if subdomainName != "" {
		newURL := fmt.Sprintf("https://hkt.sh/%v%v", subdomainName, request.Path)
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("<html>\n<head><title>hkt.sh</title></head>\n<body><a href=\"%v\">moved here</a></body>\n</html>", newURL),
			StatusCode: 301,
			Headers: map[string]string{
				"Location":      newURL,
				"Cache-Control": "private, max-age=90",
			},
		}, nil
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

	// Substitute {param} placeholder if additional path segments are provided
	redirectURL := entry.URL
	rawParam := request.PathParameters["param"]
	if rawParam != "" {
		param, err := url.PathUnescape(rawParam)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		redirectURL = strings.ReplaceAll(redirectURL, "{param}", param)
	}

	_, err = svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("hkt-sh-entries"),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String(name),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":inc": {
				N: aws.String("1"),
			},
		},
		UpdateExpression: aws.String("ADD AccessCount :inc"),
	})

	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("<html>\n<head><title>hkt.sh</title></head>\n<body><a href=\"%v\">moved here</a></body>\n</html>", redirectURL),
		StatusCode: 301,
		Headers: map[string]string{
			"Location":      redirectURL,
			"Cache-Control": "private, max-age=90",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
