FROM golang:latest AS build

ADD . /go/src/github.com/usrpro/file-handling-api

WORKDIR /go/src/github.com/usrpro/file-handling-api

RUN go get ./... && go build -v 

################################################
FROM debian:stretch-slim

RUN apt-get -y update && apt-get -y install ca-certificates

COPY --from=build /go/src/github.com/usrpro/file-handling-api/file-handling-api /file-handling-api
COPY sql/ /sql
COPY config.json /config.json

EXPOSE 80

CMD ["/file-handling-api", "-config", "/config.json"]