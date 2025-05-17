FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
COPY proto ./proto/
COPY client ./client/

RUN go mod tidy && go mod download

RUN go build -o /client ./client

EXPOSE 50052

CMD ["/client"] 