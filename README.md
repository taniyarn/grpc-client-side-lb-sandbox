# gRPC Client-Side Load Balancing Sandbox

このプロジェクトは、gRPC のクライアントサイドロードバランシングの動作を検証するためのサンドボックス環境です。Kubernetes の ClusterIP None（ヘッドレスサービス）の挙動を Docker 環境でシミュレートし、ロードバランシングやサーバー障害からの復旧をテストできます。

## アーキテクチャ

- **サーバー**: 複数の gRPC サーバーインスタンス（server1、server2、server3）が同じ名前（server）で登録されています
- **プロキシ**: クライアントリクエストを複数のサーバーに分散する gRPC クライアント
- **ロードバランサー**: gRPC クライアント内蔵の round_robin ポリシーによるクライアントサイドロードバランシング

## 特徴

- **クライアントサイドロードバランシング**: gRPC のネイティブロードバランシング機能を使用
- **ヘッドレスサービスシミュレーション**: Kubernetes の ClusterIP None に似た動作を Docker 環境で再現
- **自動フェイルオーバー**: サーバーインスタンスの障害が発生しても自動的に別のサーバーに切り替え
- **ローリングアップデートテスト**: サーバー再起動時の接続継続性をテスト

## プロジェクト構造

```
.
├── proto/              # Protocol Buffers定義ファイル
├── server/             # サーバー側の実装
├── client/             # クライアント側の実装（プロキシ）
├── docker-compose.yml  # Docker環境定義
├── Taskfile.yml        # タスク定義
├── .golangci.yml       # golangci-lint設定ファイル
└── Dockerfile          # Dockerイメージ定義
```

## 必要条件

- Docker と Docker Compose
- Task (タスクランナー)
- Go 1.16 以上 (ローカルでビルドする場合)
- gRPCurl (テスト用)

## セットアップと実行

### Docker 環境での実行（推奨）

1. 環境を起動:

```bash
task up
```

2. プロキシを通じてサーバーにリクエストを送信:

```bash
grpcurl -plaintext -d '{"name": "World"}' localhost:50050 hello.HelloService/SayHello
```

### コード品質チェック

golangci-lint を使用してコード品質をチェックします:

```bash
# 標準のlintチェック
task lint

# 自動修正を試みる
task lint:fix

# Docker環境内でlintを実行
task lint:docker
```

### ロードバランシングのテスト

サーバー間でのリクエスト分散を確認:

```bash
task verify-lb
```

### ロードバランシングの詳細チェック

各サーバーからのレスポンスを確認:

```bash
task verify-lb-response
```

### ローリングアップデートのシミュレーション

サーバーを順番に再起動し、接続の継続性をテスト:

```bash
task simulate-rolling-update
```

### ログの確認

プロキシ（クライアント）のログを確認:

```bash
task logs:proxy
```

サーバーのログを確認:

```bash
task logs:server
```

## gRPC クライアントの詳細ログ

詳細な gRPC ログを有効にするには:

```bash
GRPC_GO_LOG_VERBOSITY_LEVEL=99 GRPC_GO_LOG_SEVERITY_LEVEL=INFO task logs:proxy
```

## Docker 環境の停止

```bash
task down
```

## 設計の詳細

- **DNS 解決**: Docker Compose のネットワークエイリアスを使用して、同じホスト名（server）に対して複数の IP アドレスが解決されるようにしています
- **接続再試行**: `WaitForReady(true)` 設定により、接続が確立するまでリクエストがブロックされます
- **ラウンドロビン**: 設定 `{"loadBalancingPolicy":"round_robin"}` により、利用可能なサーバー間でリクエストが分散されます

## トラブルシューティング

- **接続タイムアウト**: すべてのサーバーが利用不可能な場合は、Dialer のタイムアウト（5 秒）でエラーが発生します
- **サーバー障害**: 一部のサーバーが停止しても、他のサーバーが利用可能であればサービスは継続します
