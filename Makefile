.PHONY: vendor_clean fmt build clean install

OUT = ./bin/dircast

GOPATH := ${PWD}/vendor:${GOPATH}
export GOPATH

prefix=/usr/local

default: build

vendor_clean:
	rm -dRf ./vendor/

vendor:
	GOPATH=${PWD}/vendor go get -d -u -v \
				 github.com/mikkyang/id3-go \
				 gopkg.in/alecthomas/kingpin.v1

fmt: dircast.go
	go fmt dircast.go

fmt_test:
	go fmt dircast_test.go

test:
	go test .

build: vendor dircast.go
	go build -v -o $(OUT) dircast.go

clean:
	rm -dRf ./bin

install: $(OUT)
	install -m 0755 $(OUT) $(prefix)/bin
