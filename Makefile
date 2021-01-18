GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
VERSION := $(shell date -u +"%Y.%m.%d-%s")
BINARY_NAME=cloudsort

export VERSION
all: build

proto:
	protoc -I pb/ --go_out=pb/ --go_opt=paths=source_relative --go-grpc_out=pb/ --go-grpc_opt=paths=source_relative service.proto data.proto

build:
	$(GOBUILD) -ldflags '-X main.Version=$(VERSION)' -o $(BINARY_NAME) .

test: 
	$(GOTEST) -v ./...

clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

package: build
	nfpm pkg -p deb
