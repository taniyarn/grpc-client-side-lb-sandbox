# gRPC Client-Side Load Balancing Sandbox

このプロジェクトは、gRPC のクライアントサイドロードバランシングの動作を検証するためのサンドボックス環境です。

## プロジェクト構造

```
.
├── proto/                    # Protocol Buffers定義ファイル
├── server/                   # サーバー側の実装
│   └── main.go              # メインサーバー
├── client/                   # クライアント側の実装
│   └── main.go              # メインクライアント
└── go.mod                    # Goモジュール定義
```

## 必要条件

- Go 1.16 以上
- Protocol Buffers Compiler (protoc)
- Go Protocol Buffers Plugin (protoc-gen-go)

## セットアップ

1. 依存関係のインストール:

```bash
go mod download
```

2. Protocol Buffers のコンパイル:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/*.proto
```

## 使用方法

1. サーバーの起動:

```bash
go run server/main.go
```

2. クライアントの実行:

```bash
go run client/main.go
```
