.PHONY: deps clean build

deps:
	go get -u -v github.com/aws/aws-lambda-go/events
	go get -u -v github.com/aws/aws-lambda-go/lambda

clean: 
	rm -rf ./home/home

build:
	GOOS=linux GOARCH=amd64 go build -o home/home ./home
	${GOPATH}/bin/build-lambda-zip.exe --output home/home.zip home/home