# stage: builder
FROM golang:alpine AS builder

COPY . /go/src/github.com/concourse/docker-image-resource
ENV CGO_ENABLED 0
COPY assets/ /assets
RUN go build -o /assets/check github.com/concourse/docker-image-resource/cmd/check
RUN go build -o /assets/print-metadata github.com/concourse/docker-image-resource/cmd/print-metadata
RUN go build -o /assets/ecr-login github.com/concourse/docker-image-resource/vendor/github.com/awslabs/amazon-ecr-credential-helper/ecr-login/cmd
RUN set -e; \
    for pkg in $(go list ./...); do \
      go test -o "/tests/$(basename $pkg).test" -c $pkg; \
    done

# stage: resource
FROM alpine:edge AS resource
RUN apk --no-cache add \
      bash \
      docker \
      jq \
      ca-certificates \
      xz \
    ;
COPY --from=builder /assets /opt/resource
RUN ln -s /opt/resource/ecr-login /usr/local/bin/docker-credential-ecr-login

# stage: tests
FROM resource AS tests
COPY --from=builder /tests /tests
ADD . /docker-image-resource
RUN set -e; \
    for test in /tests/*.test; do \
      $test -ginkgo.v; \
    done

# stage: shelltests
FROM resource AS shelltests
RUN apk update && apk add libressl
ADD . /docker-image-resource
RUN wget --directory-prefix  /docker-image-resource/tests/shell-tests/ https://raw.githubusercontent.com/kward/shunit2/v2.1.7/shunit2
RUN cd /docker-image-resource/tests/shell-tests/ && ./test.sh

# final output stage
FROM resource
