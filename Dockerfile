ARG GOLANG_VERSION=1.25.1
ARG ALPINE_VERSION=3.22

FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS build_deps

RUN apk add --no-cache git

WORKDIR /workspace

COPY go.mod .
COPY go.sum .

RUN go mod download

FROM build_deps AS build

COPY . .

RUN CGO_ENABLED=0 go build -o mqtt2cmd -ldflags '-w -extldflags "-static"' .

FROM alpine:${ALPINE_VERSION}

RUN apk add --no-cache ca-certificates

COPY --from=build /workspace/mqtt2cmd /usr/local/bin/mqtt2cmd

ENTRYPOINT ["/usr/local/bin/mqtt2cmd"]
