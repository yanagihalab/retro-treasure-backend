# Relic Raid VPS deployment

Relic Raid is served behind nginx under `/games/`.

This deployment is designed to coexist with the existing Tozan Todoke services on the same VPS. Do not reuse ports `8080` or `8788`, and do not expose the Relic Raid internal port directly to the internet.

## Runtime layout

- App directory: `/home/ubuntu/relic-raid`
- Binary: `/home/ubuntu/relic-raid/relic-raid`
- Environment file: `/home/ubuntu/.relic-raid/production.env`
- Persistent data: `/home/ubuntu/.relic-raid/data/state.json`
- Internal bind: `127.0.0.1:8090`
- Public URL: `https://ik1-206-76937.vs.sakura.ne.jp/games/`

## Existing services to preserve

- Tozan Todoke frontend: `127.0.0.1:8080`
- Tozan Todoke backend: `127.0.0.1:8788`
- nginx public ports: `80`, `443`

Relic Raid uses only `127.0.0.1:8090`.

## Environment

Create `/home/ubuntu/.relic-raid/production.env`:

```env
APP_NAME=relic-raid
APP_HOST=127.0.0.1
APP_PORT=8090
APP_BASE_PATH=/games
DATA_DIR=/home/ubuntu/.relic-raid/data
APP_STATE_FILE=/home/ubuntu/.relic-raid/data/state.json
```

Keep the file outside the repository and do not commit secrets.

Create directories and permissions:

```bash
mkdir -p /home/ubuntu/relic-raid /home/ubuntu/.relic-raid/data
chmod 700 /home/ubuntu/.relic-raid /home/ubuntu/.relic-raid/data
chmod 600 /home/ubuntu/.relic-raid/production.env
```

## Build locally

From the repository root:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/relic-raid-deploy/relic-raid ./cmd/server
cp deploy/relic-raid.service /tmp/relic-raid-deploy/relic-raid.service
cp deploy/nginx-games-location.conf /tmp/relic-raid-deploy/nginx-games-location.conf
```

Upload:

```bash
rsync -av /tmp/relic-raid-deploy/relic-raid \
  /tmp/relic-raid-deploy/relic-raid.service \
  /tmp/relic-raid-deploy/nginx-games-location.conf \
  sakura-tozantodoke:/home/ubuntu/relic-raid/
```

## Install or update

```bash
sudo cp /home/ubuntu/relic-raid/relic-raid.service /etc/systemd/system/relic-raid.service
sudo systemctl daemon-reload
sudo systemctl enable --now relic-raid.service

sudo cp /home/ubuntu/relic-raid/nginx-games-location.conf /etc/nginx/snippets/nginx-games-location.conf
sudo cp /etc/nginx/sites-available/tozantodoke-public /etc/nginx/sites-available/tozantodoke-public.bak.$(date +%Y%m%d-%H%M%S)
sudo perl -0pi -e 's|(include /etc/nginx/snippets/zero-order-forum-location.conf;)|include /etc/nginx/snippets/nginx-games-location.conf;\n\n    $1|' /etc/nginx/sites-available/tozantodoke-public
sudo nginx -t
sudo systemctl reload nginx
sudo systemctl restart relic-raid.service
```

Only insert the nginx snippet once. If it is already included, copy the snippet and reload nginx.

## Updating an already installed server

If the nginx snippet is already included, use the shorter update flow:

```bash
sudo cp /home/ubuntu/relic-raid/relic-raid.service /etc/systemd/system/relic-raid.service
sudo systemctl daemon-reload
sudo cp /home/ubuntu/relic-raid/nginx-games-location.conf /etc/nginx/snippets/nginx-games-location.conf
sudo nginx -t
sudo systemctl reload nginx
sudo systemctl restart relic-raid.service
```

## Backup state

Before risky changes, back up the state file:

```bash
mkdir -p /home/ubuntu/relic-raid-backups
cp /home/ubuntu/.relic-raid/data/state.json \
  /home/ubuntu/relic-raid-backups/state.$(date +%Y%m%d-%H%M%S).json
```

If the file does not exist yet, no users have been persisted.

## Verification

```bash
ss -ltnp | grep -E ':8090|:8080|:8788'
systemctl status relic-raid.service --no-pager -l
curl -I http://127.0.0.1:8090/health
curl -I http://127.0.0.1:8090/games/
curl -I http://127.0.0.1:8090/games/static/js/app.js
curl http://127.0.0.1:8090/games/api/checkpoints/master | head -c 300
curl -k -I https://ik1-206-76937.vs.sakura.ne.jp/games/
curl -k -I https://ik1-206-76937.vs.sakura.ne.jp/
```

## Expected headers

`/games/` should return `200 OK` from nginx after the snippet is included and the service is restarted.

`/` should still return the existing Tozan Todoke frontend.

## Rollback

Restore the previous binary or nginx site backup, then reload:

```bash
sudo cp /etc/nginx/sites-available/tozantodoke-public.bak.<timestamp> /etc/nginx/sites-available/tozantodoke-public
sudo nginx -t
sudo systemctl reload nginx
sudo systemctl restart relic-raid.service
```
