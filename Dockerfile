FROM ubuntu:14.04

ENV GOLANG_VERSION 1.14
ENV PATH /usr/local/go/bin:$PATH
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV GOFLAGS -p=8
ENV PATH /go/bin:$PATH

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      curl \
      git \
      build-essential \
    && rm -rf /var/lib/apt/lists/* \
    && curl -sSL https://storage.googleapis.com/golang/go$GOLANG_VERSION.linux-amd64.tar.gz | tar -v -C /usr/local -xz \
	&& mkdir -p /go/src /go/bin && chmod -R 777 /go

ADD . /app/
WORKDIR /app/

RUN make \
    && make install