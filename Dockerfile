FROM concourse/busyboxplus:iptables

RUN for cert in `ls -1 /etc/ssl/certs/*.pem`; \
      do cat "$cert" >> /etc/ssl/certs/ca-certificates.crt; \
    done

ADD https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 /usr/local/bin/jq
RUN chmod +x /usr/local/bin/jq

ADD https://get.docker.io/builds/Linux/x86_64/docker-latest.tgz /usr/local/bin/docker
RUN chmod +x /usr/local/bin/docker

RUN /usr/local/bin/docker --version

ADD assets/ /opt/resource/
RUN chmod +x /opt/resource/*

ADD bin/ /bin/
