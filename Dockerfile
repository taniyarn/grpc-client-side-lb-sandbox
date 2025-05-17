FROM golang:1.24-alpine as builder

RUN apk add --no-cache protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2

WORKDIR /workspace 