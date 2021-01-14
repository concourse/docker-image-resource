ARG base_image=alpine:latest
ARG builder_image=concourse/golang-builder

# stage: builder
FROM ${builder_image} AS builder

WORKDIR /concourse/docker-image-resource
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

ENV CGO_ENABLED 0
COPY assets/ /assets
RUN go build -o /assets/check github.com/concourse/docker-image-resource/cmd/check
RUN go build -o /assets/print-metadata github.com/concourse/docker-image-resource/cmd/print-metadata
RUN go build -o /assets/ecr-login github.com/awslabs/amazon-ecr-credential-helper/ecr-login/cmd
RUN set -e; \
    for pkg in $(go list ./...); do \
      go test -o "/tests/$(basename $pkg).test" -c $pkg; \
    done

# stage: resource
FROM ${base_image} AS resource
RUN apk --no-cache add \
      bash \
      docker \
      jq \
      ca-certificates \
      xz \
      util-linux \
      tar \
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

# final output stage
FROM resource
