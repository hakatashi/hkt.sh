package main

import (
	"bytes"
	"encoding/json"
	"html/template"

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

type HomeTemplateParams struct {
	Entries []Entry
	Headers string
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	body, err := json.MarshalIndent(request.Headers, "", "\t")
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	sess, err := session.NewSession()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	svc := dynamodb.New(sess)

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String("hkt-sh-entries"),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	entries := []Entry{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &entries)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	tpl, err := template.ParseFiles("home.html.tpl")
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	var output bytes.Buffer
	err = tpl.ExecuteTemplate(&output, "home.html.tpl", HomeTemplateParams{
		Headers: string(body),
		Entries: entries,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		Body:       output.String(),
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/html; charset=utf-8",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
