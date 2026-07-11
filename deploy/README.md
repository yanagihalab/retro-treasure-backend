# Relic Raid VPS deployment

Relic Raid is served behind nginx under `/games/`.

## Runtime layout

- App directory: `/home/ubuntu/relic-raid`
- Binary: `/home/ubuntu/relic-raid/relic-raid`
- Environment file: `/home/ubuntu/.relic-raid/production.env`
- Persistent data: `/home/ubuntu/.relic-raid/data/state.json`
- Internal bind: `127.0.0.1:8090`
- Public URL: `https://ik1-206-76937.vs.sakura.ne.jp/games/`

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

## Install or update

```bash
mkdir -p /home/ubuntu/relic-raid /home/ubuntu/.relic-raid/data
chmod 700 /home/ubuntu/.relic-raid /home/ubuntu/.relic-raid/data

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

## Verification

```bash
ss -ltnp | grep -E ':8090|:8080|:8788'
systemctl status relic-raid.service --no-pager -l
curl -I http://127.0.0.1:8090/health
curl -I http://127.0.0.1:8090/games/
curl -k -I https://ik1-206-76937.vs.sakura.ne.jp/games/
curl -k -I https://ik1-206-76937.vs.sakura.ne.jp/
```
