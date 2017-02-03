# G53
[![Build Status](https://travis-ci.org/hawkingrei/G53.svg?branch=master)](https://travis-ci.org/hawkingrei/G53)
[![Build Status](https://ci.appveyor.com/api/projects/status/github/hawkingrei/g53?branch=master&svg=true)](https://ci.appveyor.com/project/hawkingrei/g53/branch/master)
[![Coverage Status](https://coveralls.io/repos/github/hawkingrei/G53/badge.svg?branch=master)](https://coveralls.io/github/hawkingrei/G53?branch=master)
[![codebeat badge](https://codebeat.co/badges/cc33aba7-de9f-4cfc-95cf-8407baddb063)](https://codebeat.co/projects/github-com-hawkingrei-g53)
[![Go Report Card](https://goreportcard.com/badge/github.com/hawkingrei/g53)](https://goreportcard.com/report/github.com/hawkingrei/g53)

#### Build

##### Building without docker:

```
export GOPATH=/tmp/go
export PATH=${PATH}:${GOPATH}/bin
go get -v github.com/tools/godep
go get -d -v github.com/hawkingrei/G53
cd ${GOPATH}/src/github.com/hawkingrei/G53
godep restore
cd ${GOPATH}/src/github.com/hawkingrei/G53/
go build -o ${GOPATH}/bin/G53
```

##### Building with docker:

```
wget https://raw.githubusercontent.com/hawkingrei/G53/master/Dockerfile
sudo docker build -t g53 .
sudo docker run -p 53:53/udp  g53
```

#### To do
- Support tls
- Add cache
