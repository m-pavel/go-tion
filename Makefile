CGO_CFLAGS="-I${PWD}/../go-gattlib/gattlib/include"
CGO_LDFLAGS="-L${PWD}/../go-gattlib/gattlib/build/dbus -lm -lutil"
GF=CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} LD_LIBRARY_PATH=${PWD}/../go-gattlib/gattlib/build/dbus

all: test build

deps:
	${GF} go get -v -d ./...
test: deps
	${GF} go test -v $$(go list ./... | grep -v /vendor/)

build-cli: deps
	${GF} go build -o tion-cli ./cli

build-influx: deps
	${GF} go build -o tion-influx ./influx

build-mqtt: deps
	${GF} go build -o tion-mqtt ./mqtt

build-schedule: deps
	${GF} go build -o tion-schedule ./schedule

build: build-cli build-influx build-schedule build-mqtt

clean: gattlib-clean
	rm -f ./tion-influx
	rm -f ./tion-schedule
	rm -f ./tion-cli
	rm -f ./tion-mqtt

