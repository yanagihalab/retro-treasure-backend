# RELIC RAID

RELIC RAID は、防災用品をキャラクター化したカードでクトゥルフモチーフのボスに挑む、横画面向けWebゲームです。GoサーバがAPIと埋め込み静的ファイルを配信します。

## 主な機能

- ユーザー登録、ログイン、トークン認証
- 心・技・体のキャラクターカード、6枚デッキ、ガチャ
- 30体のボス、耐久戦、ボス報酬
- 実地図上のチェックポイントとQR報酬
- カード図鑑とカード詳細
- `/games/` など任意のbase pathでの公開
- JSONファイルまたはMariaDBによる進行状態の永続化

## 必要環境

- Go 1.23以上
- MariaDB 10.11以上（本番運用時）

## ローカル起動

MariaDBを使わず、JSONファイルへ状態を保存する場合:

```bash
mkdir -p .local-data
APP_HOST=127.0.0.1 \
APP_PORT=8080 \
APP_BASE_PATH=/games \
DATA_DIR=.local-data \
go run ./cmd/server
```

確認先:

- ゲーム: `http://127.0.0.1:8080/games/`
- ヘルスチェック: `http://127.0.0.1:8080/health`
- チェックポイントAPI: `http://127.0.0.1:8080/games/api/checkpoints/master`

`APP_BASE_PATH` を空にすると、ゲームは `/` と `/static/` で配信されます。

## MariaDB接続

本番では、ユーザー・カード・進行状況を `relic_raid_state` テーブルのJSONとして保存します。JSON形式のため、既存のゲームモデルを欠落なく保持しながらMariaDBの `JSON_EXTRACT` でも参照できます。

MariaDB側で専用DBと最小権限ユーザーを作成します。パスワードは実際のランダム値に置き換え、リポジトリへ保存しないでください。

```sql
CREATE DATABASE relic_raid
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

CREATE USER 'relic_raid_app'@'127.0.0.1'
    IDENTIFIED BY '<strong-random-password>';

GRANT SELECT, INSERT, UPDATE, CREATE
    ON relic_raid.*
    TO 'relic_raid_app'@'127.0.0.1';
```

接続環境変数:

```env
DB_ENABLED=true
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=relic_raid
DB_USER=relic_raid_app
DB_PASSWORD=<strong-random-password>
DB_STATE_KEY=primary
```

起動時にテーブルがない場合は自動作成されます。MariaDB内に状態がなく、`APP_STATE_FILE` に既存のJSONがある場合は、その内容を一度だけMariaDBへ取り込みます。以後はMariaDBを正本として読み書きします。

手動でテーブルを作る場合は [migrations/mariadb/001_state.sql](migrations/mariadb/001_state.sql) を使用できます。

状態確認例:

```sql
SELECT state_key, updated_at, JSON_LENGTH(payload, '$.users_by_id') AS users
FROM relic_raid_state;
```

## 環境変数

| 変数 | 既定値 | 用途 |
| --- | --- | --- |
| `APP_NAME` | `retro-treasure-api` | ログに表示するサービス名 |
| `APP_HOST` | 空 | bind先。本番は `127.0.0.1` |
| `APP_PORT` | `8080` | HTTPポート |
| `APP_BASE_PATH` | 空 | 公開パス。本番は `/games` |
| `DATA_DIR` | 空 | JSON保存ディレクトリ |
| `APP_STATE_FILE` | 空 | JSON状態ファイルの明示パス |
| `DB_ENABLED` | `false` | MariaDB永続化の有効化 |
| `DB_HOST` | 空 | MariaDBホスト |
| `DB_PORT` | `3306` | MariaDBポート |
| `DB_NAME` | 空 | MariaDBデータベース名 |
| `DB_USER` | 空 | MariaDBユーザー |
| `DB_PASSWORD` | 空 | MariaDBパスワード |
| `DB_STATE_KEY` | `primary` | 状態行を識別するキー |

## API

base pathを指定した場合、次の `/api/...` の前に `/games` が付きます。

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/player/me`
- `POST /api/login-bonus/claim`
- `GET /api/notices`
- `GET /api/cards/deck`
- `GET /api/cards/collection`
- `GET /api/cards/archive`
- `POST /api/cards/deck`
- `POST /api/gacha/draw`
- `GET /api/boss`
- `POST /api/boss/auto`
- `POST /api/boss/reward`
- `GET /api/checkpoints/master`
- `GET /api/checkpoints/history`
- `POST /api/checkpoints/claim`
- `GET /health`

## テスト

```bash
go test ./...
```

VPSにはGitHubからソースコードをcloneし、`/home/ubuntu/relic-raid-source` から `go run` で起動します。サーバ上でソースを確認・編集でき、バイナリ配布は行いません。

## ディレクトリ

```text
cmd/server/                 HTTPサーバ起動処理
deploy/                     systemd・nginx・VPS手順
internal/config/            環境変数
internal/handler/           HTTPハンドラ
internal/middleware/        認証ミドルウェア
internal/model/             API・ドメインモデル
internal/repository/        メモリ状態と永続化ストア
internal/seed/              カード・ボス・チェックポイント
internal/service/           ゲームロジック
internal/webassets/static/  HTML・CSS・JavaScript・画像
migrations/mariadb/         MariaDBスキーマ
```

VPS公開の詳細は [deploy/README.md](deploy/README.md) を参照してください。
