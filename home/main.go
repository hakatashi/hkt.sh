package main

import (
	"bytes"
	"encoding/json"
	"html/template"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var ()

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	body, err := json.MarshalIndent(request.Headers, "", "\t")
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	tpl, err := template.ParseFiles("home.html.tpl")
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	var output bytes.Buffer
	err = tpl.ExecuteTemplate(&output, "home.html.tpl", map[string]string{
		"headers": string(body),
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
