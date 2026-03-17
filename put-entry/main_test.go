package main

import (
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hakatashi/hkt.sh/testutil"
)

const testTableName = "hkt-sh-entries-put-entry-test"

var testDB *dynamodb.DynamoDB

func TestMain(m *testing.M) {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}
	testDB = testutil.NewDynamoDBClient(endpoint)
	if err := testutil.CreateTable(testDB, testTableName); err != nil {
		panic(err)
	}
	code := m.Run()
	testutil.DeleteTable(testDB, testTableName)
	os.Exit(code)
}

func newTestApp() *App {
	return &App{db: testDB, tableName: testTableName}
}

func deleteEntry(t *testing.T, name string) {
	t.Helper()
	_, err := testDB.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(testTableName),
		Key:       map[string]*dynamodb.AttributeValue{"Name": {S: aws.String(name)}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutEntryValid(t *testing.T) {
	t.Cleanup(func() { deleteEntry(t, "test-entry") })

	app := newTestApp()
	resp, err := app.handler(events.APIGatewayProxyRequest{
		Body: `{"Name":"test-entry","URL":"https://example.com"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	item, err := testDB.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(testTableName),
		Key:       map[string]*dynamodb.AttributeValue{"Name": {S: aws.String("test-entry")}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.Item == nil {
		t.Fatal("expected item to be created in DynamoDB")
	}
	if *item.Item["URL"].S != "https://example.com" {
		t.Errorf("expected URL https://example.com, got %s", *item.Item["URL"].S)
	}
	if item.Item["CreatedAt"] == nil {
		t.Error("expected CreatedAt to be set")
	}
}

func TestPutEntryEmptyName(t *testing.T) {
	app := newTestApp()
	_, err := app.handler(events.APIGatewayProxyRequest{
		Body: `{"Name":"","URL":"https://example.com"}`,
	})
	if err == nil {
		t.Error("expected error for empty Name")
	}
}

func TestPutEntryEmptyURL(t *testing.T) {
	app := newTestApp()
	_, err := app.handler(events.APIGatewayProxyRequest{
		Body: `{"Name":"test","URL":""}`,
	})
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestPutEntryInvalidJSON(t *testing.T) {
	app := newTestApp()
	_, err := app.handler(events.APIGatewayProxyRequest{
		Body: `not valid json`,
	})
	if err == nil {
		t.Error("expected error for invalid JSON body")
	}
}
