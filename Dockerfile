ARG base_image=cgr.dev/chainguard/wolfi-base
ARG builder_image=concourse/golang-builder

ARG BUILDPLATFORM
FROM --platform=${BUILDPLATFORM} ${builder_image} AS builder

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

WORKDIR /src
COPY . /src
ENV CGO_ENABLED=0
RUN go mod download
COPY assets/ /assets
RUN go build -o /assets/check ./cmd/check
RUN go build -o /assets/print-metadata ./cmd/print-metadata
RUN go build -o /assets/ecr-login github.com/awslabs/amazon-ecr-credential-helper/ecr-login/cli/docker-credential-ecr-login
RUN set -e; \
    for pkg in $(go list ./...); do \
      go test -o "/tests/$(basename $pkg).test" -c $pkg; \
    done

FROM ${base_image} AS resource
RUN apk --no-cache add \
    docker \
    docker-cli-buildx \
    jq \
    ca-certificates \
    xz \
    iproute2 \
    mount \
    umount \
    cmd:tar \
    sed

COPY --from=builder /assets /opt/resource
RUN mkdir /usr/local/bin && ln -s /opt/resource/ecr-login /usr/local/bin/docker-credential-ecr-login

FROM resource AS tests
COPY --from=builder /tests /tests
ADD . /docker-image-resource
RUN set -e; \
    for test in /tests/*.test; do \
      $test -ginkgo.v; \
    done

FROM resource
