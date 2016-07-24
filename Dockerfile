FROM concourse/buildroot:iptables

ADD docker/ /usr/local/bin/
RUN /usr/local/bin/docker --version

ADD assets/ /opt/resource/
RUN chmod +x /opt/resource/*

ADD ecr-login /usr/local/bin/
RUN chmod +x /usr/local/bin/ecr-login && \
      mkdir ~/.docker && \
      echo '{"credsStore":"ecr-login"}' >> ~/.docker/config.json
