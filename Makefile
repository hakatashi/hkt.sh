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
	@if [ -f /cygdrive/c/Users/denjj/go/bin/build-lambda-zip.exe ]; then \
		cd home && /cygdrive/c/Users/denjj/go/bin/build-lambda-zip.exe --output home.zip home home.html.tpl; \
	else \
		cd home && zip home.zip home home.html.tpl; \
	fi

entry/entry.zip: entry/main.go
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o entry/entry ./entry
	@if [ -f /cygdrive/c/Users/denjj/go/bin/build-lambda-zip.exe ]; then \
		cd entry && /cygdrive/c/Users/denjj/go/bin/build-lambda-zip.exe --output entry.zip entry; \
	else \
		cd entry && zip entry.zip entry; \
	fi

put-entry/put-entry.zip: put-entry/main.go
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o put-entry/put-entry ./put-entry
	@if [ -f /cygdrive/c/Users/denjj/go/bin/build-lambda-zip.exe ]; then \
		cd put-entry && /cygdrive/c/Users/denjj/go/bin/build-lambda-zip.exe --output put-entry.zip put-entry; \
	else \
		cd put-entry && zip put-entry.zip put-entry; \
	fi
