package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"golang.org/x/net/idna"
)

var ()

type Entry struct {
	Name           string
	UrlEncodedName string
	Url            string
	CreatedAt      int64
}

type HomeTemplateParams struct {
	Entries      []Entry
	Headers      string
	UserPoolId   string
	AssetsDomain string
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	host := request.Headers["Host"]
	if host != "hkt.sh" && !strings.HasSuffix(host, ".hkt.sh") {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "",
		}, nil
	}

	if host != "hkt.sh" {
		rawName := strings.TrimSuffix(host, ".hkt.sh")
		profile := idna.New()
		name, err := profile.ToUnicode(rawName)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		newUrl := fmt.Sprintf("https://hkt.sh/%v", url.QueryEscape(name))
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("<html>\n<head><title>hkt.sh</title></head>\n<body><a href=\"%v\">moved here</a></body>\n</html>", newUrl),
			StatusCode: 301,
			Headers: map[string]string{
				"Location":      newUrl,
				"Cache-Control": "private, max-age=90",
			},
		}, nil
	}

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

	for _, entry := range entries {
		entry.UrlEncodedName = url.QueryEscape(entry.Name)
	}

	tpl, err := template.ParseFiles("home.html.tpl")
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	var output bytes.Buffer
	err = tpl.ExecuteTemplate(&output, "home.html.tpl", HomeTemplateParams{
		Headers:      string(body),
		Entries:      entries,
		UserPoolId:   os.Getenv("AUTH_USER_POOL_CLIENT_ID"),
		AssetsDomain: os.Getenv("ASSETS_WEBSITE_DOMAIN_NAME"),
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
