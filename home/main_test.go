package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hakatashi/hkt.sh/testutil"
	"golang.org/x/net/idna"
)

const testTableName = "hkt-sh-entries-home-test"

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

func clearTable(t *testing.T) {
	t.Helper()
	out, err := testDB.Scan(&dynamodb.ScanInput{
		TableName: aws.String(testTableName),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range out.Items {
		_, err := testDB.DeleteItem(&dynamodb.DeleteItemInput{
			TableName: aws.String(testTableName),
			Key:       map[string]*dynamodb.AttributeValue{"Name": item["Name"]},
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func putEntry(t *testing.T, name, entryURL, visibility string) {
	t.Helper()
	item := map[string]*dynamodb.AttributeValue{
		"Name": {S: aws.String(name)},
		"URL":  {S: aws.String(entryURL)},
	}
	if visibility != "" {
		item["Visibility"] = &dynamodb.AttributeValue{S: aws.String(visibility)}
	}
	_, err := testDB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(testTableName),
		Item:      item,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubdomainRedirectHktSh(t *testing.T) {
	app := newTestApp()
	resp, err := app.handler(events.APIGatewayProxyRequest{
		Headers: map[string]string{"Host": "foo.hkt.sh"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 301 {
		t.Errorf("expected 301, got %d", resp.StatusCode)
	}
	if resp.Headers["Location"] != "https://hkt.sh/foo" {
		t.Errorf("expected Location https://hkt.sh/foo, got %s", resp.Headers["Location"])
	}
	if resp.Headers["Cache-Control"] != "private, max-age=90" {
		t.Errorf("expected Cache-Control header, got %s", resp.Headers["Cache-Control"])
	}
}

func TestSubdomainRedirectHktSi(t *testing.T) {
	app := newTestApp()
	resp, err := app.handler(events.APIGatewayProxyRequest{
		Headers: map[string]string{"Host": "foo.hkt.si"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 301 {
		t.Errorf("expected 301, got %d", resp.StatusCode)
	}
	if resp.Headers["Location"] != "https://hkt.sh/foo" {
		t.Errorf("expected Location https://hkt.sh/foo, got %s", resp.Headers["Location"])
	}
}

func TestSubdomainRedirectURLEncoding(t *testing.T) {
	// Compute the punycode for a Japanese string, then verify the handler
	// decodes it back and redirects to the URL-encoded Unicode form.
	original := "テスト"
	profile := idna.New()
	punycode, err := profile.ToASCII(original)
	if err != nil {
		t.Fatal(err)
	}

	app := newTestApp()
	resp, err := app.handler(events.APIGatewayProxyRequest{
		Headers: map[string]string{"Host": punycode + ".hkt.sh"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 301 {
		t.Errorf("expected 301, got %d", resp.StatusCode)
	}
	expectedURL := fmt.Sprintf("https://hkt.sh/%v", url.QueryEscape(original))
	if resp.Headers["Location"] != expectedURL {
		t.Errorf("expected Location %s, got %s", expectedURL, resp.Headers["Location"])
	}
}

func TestHktSiRootRedirect(t *testing.T) {
	app := newTestApp()
	resp, err := app.handler(events.APIGatewayProxyRequest{
		Headers: map[string]string{"Host": "hkt.si"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 301 {
		t.Errorf("expected 301, got %d", resp.StatusCode)
	}
	if resp.Headers["Location"] != "https://hkt.sh" {
		t.Errorf("expected Location https://hkt.sh, got %s", resp.Headers["Location"])
	}
}

func TestUnknownHostReturns404(t *testing.T) {
	app := newTestApp()
	resp, err := app.handler(events.APIGatewayProxyRequest{
		Headers: map[string]string{"Host": "example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHomePageListsEntries(t *testing.T) {
	clearTable(t)
	t.Cleanup(func() { clearTable(t) })

	putEntry(t, "github", "https://github.com/hakatashi", "public")
	putEntry(t, "twitter", "https://twitter.com/hakatashi", "public")
	putEntry(t, "secret", "https://example.com/secret", "unlisted")

	app := newTestApp()
	resp, err := app.handler(events.APIGatewayProxyRequest{
		Headers: map[string]string{"Host": "hkt.sh"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Body, "github") {
		t.Error("expected body to contain 'github'")
	}
	if !strings.Contains(resp.Body, "twitter") {
		t.Error("expected body to contain 'twitter'")
	}
	if strings.Contains(resp.Body, "secret") {
		t.Error("expected body to NOT contain unlisted entry 'secret'")
	}
}

func TestHomePageEntriesSortedAlphabetically(t *testing.T) {
	clearTable(t)
	t.Cleanup(func() { clearTable(t) })

	putEntry(t, "zzz", "https://example.com/zzz", "public")
	putEntry(t, "aaa", "https://example.com/aaa", "public")
	putEntry(t, "mmm", "https://example.com/mmm", "public")

	app := newTestApp()
	resp, err := app.handler(events.APIGatewayProxyRequest{
		Headers: map[string]string{"Host": "hkt.sh"},
	})
	if err != nil {
		t.Fatal(err)
	}

	aPos := strings.Index(resp.Body, "aaa")
	mPos := strings.Index(resp.Body, "mmm")
	zPos := strings.Index(resp.Body, "zzz")

	if aPos == -1 || mPos == -1 || zPos == -1 {
		t.Fatal("not all entries found in response body")
	}
	if !(aPos < mPos && mPos < zPos) {
		t.Error("entries are not sorted alphabetically")
	}
}
