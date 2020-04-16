GIT_VER := $(shell git describe --tags)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)

.PHONY: test install clean

cmd/kinesis-tailf/kinesis-tailf: *.go cmd/kinesis-tailf/*.go
	cd cmd/kinesis-tailf && go build -ldflags "-s -w -X main.version=${GIT_VER} -X main.buildDate=${DATE}" -gcflags="-trimpath=${PWD}"

install: cmd/kinesis-tailf/kinesis-tailf
	install cmd/kinesis-tailf/kinesis-tailf ${GOPATH}/bin

test:
	go test -race ./...

clean:
	rm -f cmd/kinesis-tailf/kinesis-tailf
	rm -f pkg/*
