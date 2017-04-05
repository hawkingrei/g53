# This is how we want to name the binary output
OUTPUT=main
# These are the values we want to pass for Version and BuildTime
GITTAG=`git rev-parse --short HEAD`
BUILD_TIME=`date -u +%Y.%m.%d-%H:%M:%S%Z`
VERSION=0.0.1
# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X "github.com/hawkingrei/g53/version".GitCommit=${GITTAG} -X "github.com/hawkingrei/g53/version".BuildTime=${BUILD_TIME} -X "github.com/hawkingrei/g53/version".Version=${VERSION}"
all: compile

compile:
	go build ${LDFLAGS} -o g53
docker:
	go build ${LDFLAGS} -o ${GOPATH}/bin/g53
