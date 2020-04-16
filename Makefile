CGO_CFLAGS="-I${PWD}/../go-gattlib/gattlib/include"
CGO_LDFLAGS="-L${PWD}/../go-gattlib/gattlib/build/dbus -lm -lutil"

IMPL?=muka
ARCH?=$(go env GOARCH)
TAGS=-tags ${IMPL}

GF=CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} LD_LIBRARY_PATH=${PWD}/../go-gattlib/gattlib/build/dbus GOARCH=${ARCH}
all: test build

deps:
	${GF} go get ${TAGS} -v -d ./...

test: deps
	${GF} go test ${TAGS} -v $$(go list ./... | grep -v /vendor/)

lint: deps
	${GF} ${GOPATH}/bin/golangci-lint run -v ./...
	${GF} ${GOPATH}/bin/golint  $$(go list ./... | grep -v /vendor/)

build-cli: deps
	${GF} go build -o tion-cli ${TAGS} ./cmd/cli
	${GF} go build -o tion-cli-mqtt -tags mqttcli ./cmd/cli

build-influx: deps
	${GF} go build -o tion-influx ${TAGS} ./cmd/influx

build-mqtt: deps
	${GF} go build -o tion-mqtt ${TAGS} ./cmd/mqtt

build-schedule: deps
	${GF} go build -o tion-schedule ${TAGS} ./cmd/schedule

build: build-cli build-influx build-schedule build-mqtt

clean:
	rm -f ./tion-influx
	rm -f ./tion-schedule
	rm -f ./tion-cli
	rm -f ./tion-mqtt
	rm -f ./tion-cli-mqtt
