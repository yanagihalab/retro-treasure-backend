# RELIC RAID VPSデプロイ

RELIC RAIDをGitHubからVPSへcloneし、ソースコードのまま `/games/` で公開します。実行元は `/home/ubuntu/relic-raid-source` です。VPSへビルド済みバイナリは配布しません。

## 構成

| 項目 | 値 |
| --- | --- |
| 公開URL | `https://ik1-206-76937.vs.sakura.ne.jp/games/` |
| Git作業ツリー | `/home/ubuntu/relic-raid-source` |
| Go | `/home/ubuntu/.local/go/bin/go` |
| 内部bind | `127.0.0.1:8090` |
| systemd | `relic-raid.service` |
| 環境変数 | `/home/ubuntu/.relic-raid/production.env` |
| MariaDB | `127.0.0.1:3306/relic_raid` |
| nginx snippet | `/etc/nginx/snippets/nginx-games-location.conf` |

既存のTozan Todokeが使う `127.0.0.1:8080` と `127.0.0.1:8788` は変更しません。RELIC RAIDの `8090` とMariaDBの `3306` は外部公開しません。

## 初回配置

VPSへ接続してGitHubからcloneします。

```bash
ssh sakura-tozantodoke
git clone https://github.com/yanagihalab/retro-treasure-backend.git \
  /home/ubuntu/relic-raid-source
cd /home/ubuntu/relic-raid-source
```

Goをubuntuユーザー専用領域へ導入します。スクリプトは配布元とSHA-256を固定し、`sudo` を使いません。

```bash
chmod 755 deploy/install-go-user.sh
./deploy/install-go-user.sh
/home/ubuntu/.local/go/bin/go version
```

## MariaDB初期設定

MariaDBがVPSにインストール済みであることを確認してから実行します。

```bash
chmod 755 deploy/setup-mariadb.sh
./deploy/setup-mariadb.sh
```

このスクリプトは次を行います。

- `relic_raid` データベースを作成
- `relic_raid_app@127.0.0.1` に必要最小限の権限を付与
- ランダムなDBパスワードを生成
- `/home/ubuntu/.relic-raid/production.env` をモード `600` で作成

DBパスワードは画面へ表示せず、Gitにも保存しません。既存の環境ファイルがある場合は日時付きでバックアップします。

起動時にMariaDBの状態テーブルが空で、旧 `/home/ubuntu/.relic-raid/data/state.json` が存在する場合、その内容を一度だけMariaDBへ取り込みます。移行確認が終わるまで旧JSONは削除しません。

## nginx include

既存Tozan Todokeのserver blockには、次のincludeが1回だけ必要です。

```nginx
include /etc/nginx/snippets/nginx-games-location.conf;
```

未設定の場合だけ、既存設定をバックアップしてから追加してください。既存のserver blockやlocationは置き換えません。

```bash
sudo cp /etc/nginx/sites-available/tozantodoke-public \
  /etc/nginx/sites-available/tozantodoke-public.bak.$(date +%Y%m%d-%H%M%S)
```

includeを追加した後は、必ず検証してからreloadします。

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## 初回起動

ソース、systemd、nginx、MariaDBの準備をまとめて検証し、サービスを切り替えます。

```bash
cd /home/ubuntu/relic-raid-source
chmod 755 deploy/deploy-source-vps.sh
./deploy/deploy-source-vps.sh
```

このスクリプトは `go test ./...`、設定バックアップ、`nginx -t`、systemd再起動、内部・外部URLの疎通確認を順に実行します。切り替え後の検証に失敗した場合は、直前のsystemd/nginx設定へ戻します。

systemdは次のソースを直接起動します。

```ini
WorkingDirectory=/home/ubuntu/relic-raid-source
ExecStart=/home/ubuntu/.local/go/bin/go run -mod=readonly ./cmd/server
```

## GitHubから更新

VPSの作業ツリーがcleanなことを確認してから、fast-forwardで更新します。

```bash
cd /home/ubuntu/relic-raid-source
git status --short
git pull --ff-only origin main
/home/ubuntu/.local/go/bin/go test ./...
./deploy/deploy-source-vps.sh
```

## VPS上で編集

ソースは通常のGit作業ツリーなので、SSH接続後に直接確認・編集できます。

```bash
cd /home/ubuntu/relic-raid-source
git status
git diff
```

編集後はテストしてから再起動します。

```bash
/home/ubuntu/.local/go/bin/go test ./...
sudo systemctl restart relic-raid.service
journalctl -u relic-raid.service -n 100 --no-pager
```

VPSだけにある変更を失わないよう、`git pull` 前にcommitしてGitHubへpushするか、別ブランチへ退避してください。`production.env`、MariaDBデータ、進行状態はGitへ追加しません。

## 動作確認

```bash
ss -ltnp | grep -E ':3306|:8090|:8080|:8788'
systemctl status mariadb.service relic-raid.service --no-pager -l

curl -i http://127.0.0.1:8090/health
curl -I http://127.0.0.1:8090/games/
curl -I http://127.0.0.1:8090/games/static/js/transitions.js
curl http://127.0.0.1:8090/games/api/checkpoints/master | head -c 300

curl -I https://ik1-206-76937.vs.sakura.ne.jp/games/
curl -I https://ik1-206-76937.vs.sakura.ne.jp/
```

MariaDB状態確認:

```bash
sudo mariadb relic_raid -e \
  "SELECT state_key, updated_at, JSON_LENGTH(payload, '$.users_by_id') AS users FROM relic_raid_state;"
```

`/health` が `503` の場合は、MariaDB接続、権限、`production.env` を確認します。正常時は `{"status":"ok"}` を返します。

## 運用コマンド

```bash
sudo systemctl restart relic-raid.service
sudo systemctl stop relic-raid.service
sudo systemctl start relic-raid.service
systemctl status relic-raid.service --no-pager -l
journalctl -u relic-raid.service -n 100 --no-pager
```

サービスが実際にソースから起動していることは次で確認できます。

```bash
systemctl show relic-raid.service -p WorkingDirectory -p ExecStart
git -C /home/ubuntu/relic-raid-source rev-parse --short HEAD
```
