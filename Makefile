# This is how we want to name the binary output
OUTPUT=main
# These are the values we want to pass for Version and BuildTime
GITTAG=`git rev-parse --short HEAD`
BUILD_TIME=`date +%FT%T%z`
VERSION=0.0.1
# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X "github.com/hawkingrei/g53/utils/cmdline".GitCommit=${GITTAG} -X "github.com/hawkingrei/g53/utils/cmdline".BuildTime=${BUILD_TIME} -X "github.com/hawkingrei/g53/utils/cmdline".Version=${VERSION}"
all:
	go build ${LDFLAGS} -o ${GOPATH}/bin/g53 
