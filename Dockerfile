FROM golang:alpine AS builder

COPY . /go/src/github.com/concourse/docker-image-resource
COPY assets/ /assets

RUN export CGO_ENABLED=0 && go build -o /assets/check github.com/concourse/docker-image-resource/cmd/check
RUN export CGO_ENABLED=0 && go build -o /assets/print-metadata github.com/concourse/docker-image-resource/cmd/print-metadata
RUN export CGO_ENABLED=0 && go build -o /assets/ecr-login github.com/concourse/docker-image-resource/vendor/github.com/awslabs/amazon-ecr-credential-helper/ecr-login/cmd

RUN set -e; for pkg in $(go list ./...); do \
		go test -o "/tests/$(basename $pkg).test" -c $pkg; \
	done

FROM alpine:3.6 AS resource
RUN apk --no-cache --update add bash docker jq ca-certificates
COPY --from=builder /assets /opt/resource
RUN mv /opt/resource/ecr-login /usr/local/bin/docker-credential-ecr-login

FROM resource AS tests
COPY --from=builder /tests /tests
ADD . /docker-image-resource
RUN set -e; for test in /tests/*.test; do \
		$test -ginkgo.v; \
	done

FROM resource
