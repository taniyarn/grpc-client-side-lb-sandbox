FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod ./
COPY proto ./proto/
COPY server ./server/

RUN go mod download

RUN go build -o /server ./server

EXPOSE 50051

CMD ["/server"] 