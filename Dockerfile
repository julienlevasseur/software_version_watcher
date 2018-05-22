FROM golang:latest

LABEL version="0.1"
LABEL description="software_version_watcher"
MAINTAINER Julien Levasseur

RUN git clone -b dev https://github.com/julienlevasseur/software_version_watcher.git \
&& cd software_version_watcher \
&& export GOBIN=${PWD} \
&& go get \
&& go build

EXPOSE 8080

WORKDIR /go/software_version_watcher
CMD "./software_version_watcher"