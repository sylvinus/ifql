FROM ubuntu:latest

# This dockerfile is capabable of performing all
# build/test/package/deploy actions needed for IFQL.

MAINTAINER support@influxdb.com

RUN apt-get -qq update && apt-get -qq install -y \
    wget \
    git \
    mercurial \
    make

# Install go
ENV GOPATH /root/go
ENV GO_VERSION 1.9.1
ENV GO_ARCH amd64
RUN wget -q https://storage.googleapis.com/golang/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz; \
   tar -C /usr/local/ -xf /go${GO_VERSION}.linux-${GO_ARCH}.tar.gz ; \
   rm /go${GO_VERSION}.linux-${GO_ARCH}.tar.gz
ENV PATH /usr/local/go/bin:$PATH

# Install go dep
RUN go get github.com/golang/dep/...

ENV PROJECT_DIR $GOPATH/src/github.com/influxdata/ifql
ENV PATH $GOPATH/bin:$PATH
RUN mkdir -p $PROJECT_DIR
WORKDIR $PROJECT_DIR

VOLUME $PROJECT_DIR
VOLUME /root/go/src
