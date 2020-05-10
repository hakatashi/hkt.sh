.PHONY: deps clean build

deps:
	go get -u -v github.com/aws/aws-lambda-go/events
	go get -u -v github.com/aws/aws-lambda-go/lambda

clean: 
	rm -rf ./hello-world/hello-world

build:
	GOOS=linux GOARCH=amd64 go build -o hello-world/hello-world ./hello-world
	${GOPATH}/bin/build-lambda-zip.exe --output hello-world/hello-world.zip hello-world/hello-world