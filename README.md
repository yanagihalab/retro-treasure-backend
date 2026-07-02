# Retro Treasure Game Backend

ガラケー時代の探索ソーシャルゲームを現代Web向けに再構成するための、Go製バックエンド雛形です。

このリポジトリは次の2つを目的にしています。

- すぐに起動して API 動作を確認できること
- 後から DB 永続化、イベント、ランキング、フレンド機能を追加しやすいこと

元作品の固有名称・画像・世界観を流用せず、**探索型コレクションゲーム**として再設計する前提です。

---


## Docker での実行

### 前提

- Docker
- Docker Compose Plugin (`docker compose`)

### 1. Docker 単体で起動

ビルド:

```bash
docker build -t retro-treasure-backend .
```

起動:

```bash
docker run --rm -p 8080:8080   -e APP_NAME=retro-treasure-api   -e APP_PORT=8080   retro-treasure-backend
```

ヘルスチェック:

```bash
curl http://localhost:8080/health
```

期待される返却:

```json
{"status":"ok"}
```

### 2. docker compose で起動

```bash
docker compose up --build
```

バックグラウンド起動:

```bash
docker compose up -d --build
```

停止:

```bash
docker compose down
```

### 3. API 動作確認例

ユーザー登録:

```bash
curl -X POST http://localhost:8080/api/auth/register   -H "Content-Type: application/json"   -d '{
    "username":"player1",
    "password":"password123"
  }'
```

ログイン:

```bash
curl -X POST http://localhost:8080/api/auth/login   -H "Content-Type: application/json"   -d '{
    "username":"player1",
    "password":"password123"
  }'
```

返却された `token` を使ってプレイヤー情報取得:

```bash
curl http://localhost:8080/api/player/me   -H "Authorization: Bearer <token>"
```

### 4. ポート変更

ホスト側の公開ポートを 18080 にしたい場合:

```yaml
services:
  backend:
    ports:
      - "18080:8080"
```

### 5. 現状の注意点

- 現在は **インメモリ実装** です
- コンテナ再起動で登録ユーザーや進行状態は消えます
- 永続化したい場合は次段階で PostgreSQL 対応へ切り替えます

---

## Android アプリでの実行

Android Studio で `android/` ディレクトリを開くと、RELIC RAID の WebView アプリとして実行できます。

開発時は先にバックエンドを起動してください。

```bash
go run ./cmd/server
```

Android Emulator では初期設定のまま `http://10.0.2.2:8080/static/` を読み込みます。

実機で確認する場合は `android/local.properties.example` を `android/local.properties` にコピーして、Mac の LAN IP を指定してください。

```properties
GAME_BASE_URL=http://<your-mac-lan-ip>:8080/static/
```

詳細は `android/README.md` を参照してください。

---

## 1. 企画概要

### 1.1 コンセプト

短時間で繰り返し遊べる探索ループを中心に、次の体験を提供します。

- スタミナ消費による軽い周回
- 宝物や素材の収集
- 図鑑の登録と収集率上昇
- ログインボーナスによる日次継続
- 将来的なイベント拡張

### 1.2 想定プラットフォーム

- PC ブラウザ
- スマホブラウザ
- 将来的に PWA 対応可能

### 1.3 MVP の範囲

- ユーザー登録 / ログイン
- マイページ情報取得
- エリア一覧取得
- 探索実行
- 所持品一覧取得
- 図鑑一覧取得
- ログインボーナス受け取り
- お知らせ一覧取得

---

## 2. ゲーム仕様

### 2.1 基本ループ

1. ユーザーがマイページを開く
2. スタミナ残量を確認する
3. 探索エリアを選ぶ
4. 探索結果を受け取る
5. アイテム / コイン / 経験値を獲得する
6. 新規アイテムなら図鑑に登録される
7. レベルアップや新エリア開放へつながる

### 2.2 プレイヤーパラメータ

- レベル
- 経験値
- スタミナ
- 最大スタミナ
- コイン
- ジェム
- 総探索回数

### 2.3 探索の結果タイプ

- `item`: アイテム発見
- `nothing`: 空振り
- `rare_chest`: レア宝箱
- `enemy`: 敵遭遇

### 2.4 初期ルール

- スタミナは 5 分で 1 回復
- レベルアップ時に最大スタミナ +2
- エリアごとにスタミナ消費量が異なる
- 探索報酬はサーバ側で決定する

### 2.5 レベル設計例

- Lv1: 0 EXP
- Lv2: 10 EXP
- Lv3: 30 EXP
- Lv4: 60 EXP
- Lv5: 100 EXP
- Lv6: 150 EXP
- 以降は簡易式で増加

---

## 3. API 一覧

### 認証

- `POST /api/auth/register`
- `POST /api/auth/login`

### プレイヤー

- `GET /api/player/me`

### エリア

- `GET /api/areas`

### 探索

- `POST /api/explore`

### アイテム

- `GET /api/items/inventory`
- `GET /api/encyclopedia`

### 日次

- `POST /api/login-bonus/claim`

### 運営情報

- `GET /api/notices`

### ヘルスチェック

- `GET /health`

---

## 4. 画面一覧

### 4.1 タイトル画面

- 新規登録
- ログイン
- お知らせへの導線

### 4.2 マイページ

- プレイヤー名
- レベル
- 経験値
- スタミナ
- コイン
- 各機能への導線

### 4.3 探索エリア選択

- 開放済みエリア一覧
- 必要レベル
- スタミナ消費
- エリア説明

### 4.4 探索結果画面

- 結果メッセージ
- 経験値増加
- コイン増加
- アイテム獲得
- 図鑑登録
- レベルアップ表示

### 4.5 所持品一覧

- アイテム名
- レア度
- 個数

### 4.6 図鑑画面

- 登録済みアイテム
- 未発見枠
- 収集率

### 4.7 ログインボーナス画面

- 今日の報酬
- 受取状態

### 4.8 お知らせ画面

- タイトル
- 本文
- 投稿日時

---

## 5. データモデル

本雛形は**インメモリ実装**で動作します。

ただし、将来 PostgreSQL へ移行しやすいように、リポジトリ層とマイグレーション SQL の雛形も含めています。

### 主なモデル

- `User`
- `PlayerStatus`
- `Area`
- `Item`
- `InventoryEntry`
- `EncyclopediaEntry`
- `Notice`
- `ExplorationLog`

### 永続化へ移行する際の主テーブル

- `users`
- `player_status`
- `areas`
- `items`
- `area_drop_tables`
- `user_items`
- `encyclopedia_entries`
- `exploration_logs`
- `login_bonus_logs`
- `notices`

`migrations/001_init.sql` に初期 SQL の雛形を置いています。

---

## 6. ディレクトリ構成

```text
retro-treasure-backend/
├─ cmd/server/main.go
├─ internal/
│  ├─ config/
│  ├─ handler/
│  ├─ middleware/
│  ├─ model/
│  ├─ repository/
│  ├─ seed/
│  └─ service/
├─ migrations/001_init.sql
├─ go.mod
└─ README.md
```

---

## 7. 起動方法

### 7.1 必要環境

- Go 1.22 以上

### 7.2 起動

```bash
go run ./cmd/server
```

デフォルトでは `:8080` で起動します。

```bash
curl http://localhost:8080/health
```

期待される応答:

```json
{"status":"ok"}
```

### 7.3 環境変数

- `APP_PORT`: 省略時 `8080`
- `APP_NAME`: 省略時 `retro-treasure-api`

例:

```bash
APP_PORT=9090 go run ./cmd/server
```

---

## 8. 動作確認例

### 8.1 ユーザー登録

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"player1","password":"password123"}'
```

### 8.2 ログイン

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"player1","password":"password123"}'
```

レスポンス例:

```json
{
  "token": "...",
  "user_id": 1
}
```

以降は `Authorization: Bearer <token>` を付与します。

### 8.3 マイページ情報取得

```bash
curl http://localhost:8080/api/player/me \
  -H 'Authorization: Bearer <token>'
```

### 8.4 エリア一覧取得

```bash
curl http://localhost:8080/api/areas \
  -H 'Authorization: Bearer <token>'
```

### 8.5 探索実行

```bash
curl -X POST http://localhost:8080/api/explore \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"area_id":1}'
```

### 8.6 所持品一覧

```bash
curl http://localhost:8080/api/items/inventory \
  -H 'Authorization: Bearer <token>'
```

### 8.7 図鑑一覧

```bash
curl http://localhost:8080/api/encyclopedia \
  -H 'Authorization: Bearer <token>'
```

### 8.8 ログインボーナス受け取り

```bash
curl -X POST http://localhost:8080/api/login-bonus/claim \
  -H 'Authorization: Bearer <token>'
```

### 8.9 お知らせ一覧

```bash
curl http://localhost:8080/api/notices \
  -H 'Authorization: Bearer <token>'
```

---

## 9. 実装方針

### 9.1 レイヤ分離

- `handler`: HTTP 入出力
- `service`: 業務ロジック
- `repository`: データアクセス
- `model`: ドメイン構造体

### 9.2 現在の実装

- 認証は簡易トークン方式
- データはインメモリ保持
- 再起動するとユーザーデータは消える
- 初期エリア / アイテム / お知らせは起動時にシードされる

### 9.3 将来の拡張ポイント

- PostgreSQL への置き換え
- JWT 導入
- パスワードハッシュの強化
- フレンド / ランキング / イベント追加
- 期間限定ドロップテーブル
- 管理画面追加

---

## 10. 探索ロジックの概要

探索時には次を行います。

1. ユーザー認証
2. プレイヤー状態の取得
3. 自然回復分のスタミナ反映
4. エリア開放条件チェック
5. スタミナ不足チェック
6. ドロップ候補から重み付き抽選
7. 経験値 / コイン / アイテム反映
8. 図鑑登録判定
9. レベルアップ判定
10. 探索ログ保存

---

## 11. 初期シードデータ

### エリア

- 草原の入口
- 古代遺跡
- 深緑の森

### アイテム

- きらめく石
- 古びたコイン
- 幻の羽根
- 遺跡の欠片
- 神秘の宝玉

### お知らせ

- サービス開始のお知らせ
- 探索キャンペーン準備中

---

## 12. 権利面の注意

この雛形は**復刻ドリランド風のゲーム性**を参考にした設計です。

以下は避けてください。

- 元作品の名称利用
- 元作品のキャラクター名利用
- 元画像 / 元UI の直接流用
- 固有設定の過度な再現

公開・商用化を前提とする場合は、名称・世界観・アート・文言をオリジナル化してください。

---

## 13. 次の実装候補

優先度順の候補です。

1. PostgreSQL リポジトリ実装
2. フロントエンド画面一式
3. イベントシステム追加
4. ランキング追加
5. 管理者向けお知らせ編集 API
6. Docker / docker-compose 対応

---

## 14. ライセンス

必要に応じて設定してください。社内試作や研究用なら未設定でも構いませんが、公開時は明示を推奨します。
