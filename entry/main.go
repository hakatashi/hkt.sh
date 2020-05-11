package main

import (
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var ()

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	name, ok := request.PathParameters["name"]
	if !ok {
		return events.APIGatewayProxyResponse{}, errors.New("name parameter not found")
	}

	mapping := map[string]string{
		"wca": "https://www.worldcubeassociation.org/persons/2018TAKA03",
		"pcs": "https://pink-check.school/producer/detail/151777916",
	}

	url, ok := mapping[name]
	if !ok {
		return events.APIGatewayProxyResponse{}, errors.New("proper mapping were not found")
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("<html>\n<head><title>hkt.sh</title></head>\n<body><a href=\"%v\">moved here</a></body>\n</html>", url),
		StatusCode: 301,
		Headers: map[string]string{
			"Location":      url,
			"Cache-Control": "private, max-age=90",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
