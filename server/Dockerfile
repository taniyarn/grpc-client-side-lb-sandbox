FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod ./
COPY proto ./proto/
COPY server ./server/

RUN go get google.golang.org/grpc && \
    go get google.golang.org/protobuf && \
    go mod download

RUN go build -o /server ./server

EXPOSE 50051

CMD ["/server"] 