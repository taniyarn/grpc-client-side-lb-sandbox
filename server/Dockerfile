FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
COPY proto ./proto/
COPY server ./server/

RUN go mod tidy && go mod download

RUN go build -o /server ./server

EXPOSE 50051

CMD ["/server"] 