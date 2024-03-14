GIT_VER := $(shell git describe --tags)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)

.PHONY: test install clean all

all: cmd/kinesis-tailf/kinesis-tailf cmd/kinesis-cat/kinesis-cat

cmd/kinesis-tailf/kinesis-tailf: *.go cmd/kinesis-tailf/*.go go.*
	cd cmd/kinesis-tailf && go build -ldflags "-s -w -X main.version=${GIT_VER} -X main.buildDate=${DATE}" -gcflags="-trimpath=${PWD}"

cmd/kinesis-cat/kinesis-cat: *.go cmd/kinesis-cat/*.go go.*
	cd cmd/kinesis-cat && go build -ldflags "-s -w -X main.version=${GIT_VER} -X main.buildDate=${DATE}" -gcflags="-trimpath=${PWD}"

install: cmd/kinesis-tailf/kinesis-tailf cmd/kinesis-cat/kinesis-cat
	install cmd/kinesis-tailf/kinesis-tailf cmd/kinesis-cat/kinesis-cat ${GOPATH}/bin

test:
	go test -race ./...

clean:
	rm -f cmd/kinesis-tailf/kinesis-tailf cmd/kinesis-cat/kinesis-cat
	rm -f pkg/*

gen:
	protoc --go_out=paths=source_relative:kpl/ ./kpl.proto
