FROM golang:latest AS build

ADD . /go/src/github.com/usrpro/file-handling-api

WORKDIR /go/src/github.com/usrpro/file-handling-api

RUN go get ./... && go build -v 

################################################
from debian:stretch-slim

RUN apt-get -y update && apt-get -y install ca-certificates

COPY --from=build /go/src/github.com/usrpro/file-handling-api/file-handling-api /file-handling-api
COPY sql/ /sql

EXPOSE 9090

CMD ["/file-handling-api"]