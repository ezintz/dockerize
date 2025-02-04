# Go version is also in .github/workflows/CI&CD.yml.
FROM golang:1.19.2-alpine3.16 AS builder
SHELL ["/bin/ash","-e","-o","pipefail","-x","-c"]

LABEL org.opencontainers.image.source="https://github.com/powerman/dockerize"

RUN apk add --no-cache git=~2.36.3 openssl=~1.1.1q

COPY . /src
WORKDIR /src

RUN CGO_ENABLED=0 go install -ldflags "-s -w -X 'main.ver=$(git describe --match='v*' --exact-match)'"

FROM alpine:3.16

COPY --from=builder /go/bin/dockerize /usr/local/bin

ENTRYPOINT ["dockerize"]
CMD ["--help"]
