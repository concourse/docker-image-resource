FROM ubuntu:14.04

ADD http://stedolan.github.io/jq/download/linux64/jq /usr/local/bin/jq
RUN chmod +x /usr/local/bin/jq

RUN apt-get update && apt-get -y install docker.io

ADD https://get.docker.io/builds/Linux/x86_64/docker-latest /usr/local/bin/docker
RUN chmod +x /usr/local/bin/docker
RUN ln -sf /usr/local/bin/docker /usr/bin/docker.io

ADD assets/ /opt/resource/
RUN chmod +x /opt/resource/*
