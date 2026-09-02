#!/usr/bin/env bash
set -euo pipefail

readonly source_dir="/home/ubuntu/relic-raid-source"
readonly state_file="/home/ubuntu/.relic-raid/data/state.json"
readonly timestamp="$(date +%Y%m%d-%H%M%S)"
readonly go_binary="/home/ubuntu/.local/go/bin/go"
readonly unit_file="/etc/systemd/system/relic-raid.service"
readonly nginx_file="/etc/nginx/snippets/nginx-games-location.conf"
readonly unit_backup="${unit_file}.bak.${timestamp}"
readonly nginx_backup="${nginx_file}.bak.${timestamp}"

rollback_required=false

rollback() {
  local exit_code=$?

  if [[ "${rollback_required}" == true ]]; then
    echo "Deployment verification failed. Restoring the previous systemd and nginx settings." >&2
    if sudo test -f "${unit_backup}"; then
      sudo cp "${unit_backup}" "${unit_file}"
    else
      sudo rm -f "${unit_file}"
    fi
    if sudo test -f "${nginx_backup}"; then
      sudo cp "${nginx_backup}" "${nginx_file}"
    else
      sudo rm -f "${nginx_file}"
    fi
    sudo systemctl daemon-reload
    sudo nginx -t
    sudo systemctl reload nginx
    sudo systemctl restart relic-raid.service
  fi

  exit "${exit_code}"
}

wait_for_url() {
  local url=$1
  local attempts=60

  for ((attempt = 1; attempt <= attempts; attempt++)); do
    if curl --fail --silent --show-error "${url}" >/dev/null; then
      return 0
    fi
    sleep 2
  done

  echo "Timed out waiting for ${url}." >&2
  return 1
}

if [[ ! -x "${go_binary}" ]]; then
  echo "Go is not installed at ${go_binary}." >&2
  exit 1
fi
if [[ ! -d "${source_dir}/.git" ]]; then
  echo "Git source tree not found: ${source_dir}" >&2
  exit 1
fi

if [[ -f "${state_file}" ]]; then
  cp -p "${state_file}" "${state_file}.bak.${timestamp}"
fi

cd "${source_dir}"
git status --short
"${go_binary}" test ./...

if sudo test -f "${unit_file}"; then
  sudo cp "${unit_file}" "${unit_backup}"
fi
if sudo test -f "${nginx_file}"; then
  sudo cp "${nginx_file}" "${nginx_backup}"
fi

rollback_required=true
trap rollback ERR

sudo cp "${source_dir}/deploy/relic-raid.service" "${unit_file}"
sudo systemctl daemon-reload
sudo cp "${source_dir}/deploy/nginx-games-location.conf" "${nginx_file}"
sudo nginx -t
sudo systemctl reload nginx
sudo systemctl restart relic-raid.service

wait_for_url http://127.0.0.1:8090/health
curl --fail --silent --show-error --head http://127.0.0.1:8090/games/ >/dev/null
curl --fail --silent --show-error --head https://ik1-206-76937.vs.sakura.ne.jp/games/ >/dev/null
curl --fail --silent --show-error --head https://ik1-206-76937.vs.sakura.ne.jp/ >/dev/null

rollback_required=false
trap - ERR

echo
echo "RELIC RAID is running from ${source_dir}."
