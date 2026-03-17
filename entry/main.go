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

func getEntry(svc *dynamodb.DynamoDB, name string) (*Entry, error) {
	item, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("hkt-sh-entries"),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {S: aws.String(name)},
		},
	})
	if err != nil {
		return nil, err
	}
	if item.Item == nil {
		return nil, nil
	}
	entry := &Entry{}
	if err := dynamodbattribute.UnmarshalMap(item.Item, entry); err != nil {
		return nil, err
	}
	if entry.URL == "" {
		return nil, nil
	}
	return entry, nil
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

	rawParam := request.PathParameters["param"]
	var paramValue string
	if rawParam != "" {
		paramValue, err = url.PathUnescape(rawParam)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
	}

	// Build list of candidate Names to look up, in priority order.
	// When a param segment exists, try exact match (e.g. "yt/some_id") first,
	// then fall back to the parameterized entry (e.g. "yt/{param}").
	var candidates []string
	if paramValue != "" {
		candidates = append(candidates, name+"/"+paramValue)
		candidates = append(candidates, name+"/{param}")
	} else {
		candidates = append(candidates, name)
	}

	var entry *Entry
	var foundName string
	for _, candidate := range candidates {
		entry, err = getEntry(svc, candidate)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		if entry != nil {
			foundName = candidate
			break
		}
	}

	if entry == nil {
		return events.APIGatewayProxyResponse{
			Body:       "<html><head><title>hkt.sh</title></head><body><h1>The entry you requested was not found.</h1></body></html>",
			StatusCode: 404,
			Headers: map[string]string{
				"Content-Type": "text/html; charset=utf-8",
			},
		}, nil
	}

	// Substitute {param} placeholder in redirect URL.
	// Re-encode paramValue for safe embedding in URLs (API Gateway delivers decoded path params).
	redirectURL := strings.ReplaceAll(entry.URL, "{param}", url.PathEscape(paramValue))

	_, err = svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("hkt-sh-entries"),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {S: aws.String(foundName)},
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
