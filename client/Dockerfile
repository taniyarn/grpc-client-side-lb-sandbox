FROM golang:1.24-alpine

WORKDIR /app

# 依存関係のみインストール
COPY go.mod go.sum ./
RUN go mod download

# プロキシポートを公開
EXPOSE 50050

# ソースコードはdocker-compose.ymlでボリュームマウントされる
# 実行コマンドもdocker-compose.ymlで指定される 