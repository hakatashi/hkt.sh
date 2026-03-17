package main

import (
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hakatashi/hkt.sh/testutil"
)

const testTableName = "hkt-sh-entries-entry-test"

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

func putEntry(t *testing.T, name, entryURL string) {
	t.Helper()
	_, err := testDB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(testTableName),
		Item: map[string]*dynamodb.AttributeValue{
			"Name": {S: aws.String(name)},
			"URL":  {S: aws.String(entryURL)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
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

func TestEntryRedirect(t *testing.T) {
	putEntry(t, "github", "https://github.com/hakatashi")
	t.Cleanup(func() { deleteEntry(t, "github") })

	app := newTestApp()
	resp, err := app.handler(events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"name": "github"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 301 {
		t.Errorf("expected 301, got %d", resp.StatusCode)
	}
	if resp.Headers["Location"] != "https://github.com/hakatashi" {
		t.Errorf("expected Location https://github.com/hakatashi, got %s", resp.Headers["Location"])
	}
	if resp.Headers["Cache-Control"] != "private, max-age=90" {
		t.Errorf("expected Cache-Control header, got %s", resp.Headers["Cache-Control"])
	}
}

func TestEntryNotFound(t *testing.T) {
	app := newTestApp()
	resp, err := app.handler(events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"name": "nonexistent-entry-xyz"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestEntryAccessCountIncrement(t *testing.T) {
	putEntry(t, "counter-test", "https://example.com")
	t.Cleanup(func() { deleteEntry(t, "counter-test") })

	app := newTestApp()
	for i := 0; i < 2; i++ {
		_, err := app.handler(events.APIGatewayProxyRequest{
			PathParameters: map[string]string{"name": "counter-test"},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	item, err := testDB.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(testTableName),
		Key:       map[string]*dynamodb.AttributeValue{"Name": {S: aws.String("counter-test")}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.Item["AccessCount"] == nil {
		t.Fatal("AccessCount not set")
	}
	if *item.Item["AccessCount"].N != "2" {
		t.Errorf("expected AccessCount 2, got %s", *item.Item["AccessCount"].N)
	}
}

func TestEntryURLEncodedName(t *testing.T) {
	putEntry(t, "日本語", "https://example.com/japanese")
	t.Cleanup(func() { deleteEntry(t, "日本語") })

	app := newTestApp()
	resp, err := app.handler(events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"name": "%E6%97%A5%E6%9C%AC%E8%AA%9E"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 301 {
		t.Errorf("expected 301, got %d", resp.StatusCode)
	}
	if resp.Headers["Location"] != "https://example.com/japanese" {
		t.Errorf("expected Location https://example.com/japanese, got %s", resp.Headers["Location"])
	}
}

func TestEntryMissingNameParameter(t *testing.T) {
	app := newTestApp()
	_, err := app.handler(events.APIGatewayProxyRequest{
		PathParameters: map[string]string{},
	})
	if err == nil {
		t.Error("expected error when name parameter is missing")
	}
}
