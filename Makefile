.PHONY: build fmt run test vendor_clean vendor_get

GOPATH := ${PWD}/_vendor:${GOPATH}
export GOPATH

default: build

build:
	go build -v -o ./bin/dircast dircast.go

fmt:
	go fmt dircast.go

vendor_clean:
	rm -dRf ./_vendor/src

vendor_get: vendor_clean
	GOPATH=${PWD}/_vendor go get -d -u -v \
				 github.com/mikkyang/id3-go \
				 gopkg.in/alecthomas/kingpin.v1

