FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod ./
COPY proto ./proto/
COPY client ./client/

RUN go get google.golang.org/grpc && \
    go get google.golang.org/protobuf && \
    go mod download

RUN go build -o /client ./client

EXPOSE 50052

CMD ["/client"] 