FROM ubuntu:20.04

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update
RUN apt-get -y install build-essential swig gccgo golang wget apt-file git vim
RUN /bin/yes | unminimize
WORKDIR /root
RUN mkdir -vp go/src/example
ENV GOPATH /root/go
COPY swig-example/class.cxx go/src/example/
COPY swig-example/example.h go/src/example/
COPY swig-example/example.i go/src/example/
RUN mkdir -vp /root/swig-example
COPY swig-example/runme.go /root/swig-example
RUN cd /root/go/src/example/ && swig -go -c++ -intgosize 64 -o example_wrap.cxx example.i
WORKDIR /root/swig-example
RUN go build
