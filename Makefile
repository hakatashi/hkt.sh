.PHONY: deps clean build assets deploy

deps:
	go get -u -v github.com/aws/aws-lambda-go/cmd/build-lambda-zip

clean:
	rm -rf home/home entry/entry **/*.zip

deploy: build sam-deploy assets

build: home/home.zip entry/entry.zip put-entry/put-entry.zip

sam-deploy:
	-sam.cmd deploy

assets:
	aws s3 sync --acl public-read assets s3://$(shell aws cloudformation describe-stacks --stack-name hkt-sh --query "Stacks[0].Outputs[?OutputKey=='BucketName'].OutputValue" --output text)

home/home.zip: home/main.go home/home.html.tpl
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o home/home ./home
	cd home && /cygdrive/c/Users/denjj/go/bin/build-lambda-zip.exe --output home.zip home home.html.tpl

entry/entry.zip: entry/main.go
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o entry/entry ./entry
	cd entry && /cygdrive/c/Users/denjj/go/bin/build-lambda-zip.exe --output entry.zip entry

put-entry/put-entry.zip: put-entry/main.go
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o put-entry/put-entry ./put-entry
	cd put-entry && /cygdrive/c/Users/denjj/go/bin/build-lambda-zip.exe --output put-entry.zip put-entry
