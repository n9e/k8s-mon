GOCMD=go
GOBUILD=${GOCMD} build
GOCLEAN=${GOCMD} clean
GOTEST=${GOCMD} test
GOGET=${GOCMD} get
DATE= `date +%FT%T%z`


BINARY_NAME="`pwd |awk -F '/' '{print $NF}'`"
BINARY_LINUX=${BINARY_NAME}_linux
BUILDUSER=`whoami`@`hostname`
BUILDDATE=`date  +'%Y-%m-%d %H:%M:%S'`
GITREVISION=`git rev-parse HEAD`
GITVERSION=`cat VERSION`
GITBRANCH=`git symbolic-ref --short -q HEAD`
LDFLAGES=" -X 'github.com/prometheus/common/version.BuildUser=${BUILDUSER}' -X 'github.com/prometheus/common/version.BuildDate=${BUILDDATE}' -X 'github.com/prometheus/common/version.Revision=${GITREVISION}' -X 'github.com/prometheus/common/version.Version=${GITVERSION}' -X 'github.com/prometheus/common/version.Branch=${GITBRANCH}' "

all:  build
build:
		${GOBUILD} -v  -ldflags ${LDFLAGES} -o ${BINARY_NAME}
test:
		${GOTEST} -v ./...
clean:
		${GOCLEAN}
		rm -f ${BINARY_NAME}
		rm -f ${BINARY_LINUX}
run:
		${GOBUILD} -o ${BINARY_NAME} -v ./...
		./${BINARY_NAME}


build-linux:
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${GOBUILD} -o ${BINARY_LINUX} -v