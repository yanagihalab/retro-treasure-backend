#!/usr/bin/env bash
set -euo pipefail

readonly app_state_dir="/home/ubuntu/.relic-raid"
readonly env_file="${app_state_dir}/production.env"
readonly database_name="relic_raid"
readonly database_user="relic_raid_app"

if ! command -v mariadb >/dev/null 2>&1; then
  echo "MariaDB client is not installed." >&2
  exit 1
fi
if ! command -v openssl >/dev/null 2>&1; then
  echo "OpenSSL is required to generate the database password." >&2
  exit 1
fi

database_password=$(openssl rand -hex 32)

sudo mariadb <<SQL
CREATE DATABASE IF NOT EXISTS ${database_name}
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS '${database_user}'@'127.0.0.1'
    IDENTIFIED BY '${database_password}';

ALTER USER '${database_user}'@'127.0.0.1'
    IDENTIFIED BY '${database_password}';

GRANT SELECT, INSERT, UPDATE, CREATE
    ON ${database_name}.*
    TO '${database_user}'@'127.0.0.1';

FLUSH PRIVILEGES;
SQL

install -d -m 700 "${app_state_dir}" "${app_state_dir}/data"
if [[ -f "${env_file}" ]]; then
  cp -p "${env_file}" "${env_file}.bak.$(date +%Y%m%d-%H%M%S)"
fi

umask 077
temporary_env=$(mktemp "${app_state_dir}/production.env.XXXXXX")
{
  printf '%s\n' \
    "APP_NAME=relic-raid" \
    "APP_HOST=127.0.0.1" \
    "APP_PORT=8090" \
    "APP_BASE_PATH=/games" \
    "DATA_DIR=/home/ubuntu/.relic-raid/data" \
    "APP_STATE_FILE=/home/ubuntu/.relic-raid/data/state.json" \
    "DB_ENABLED=true" \
    "DB_HOST=127.0.0.1" \
    "DB_PORT=3306" \
    "DB_NAME=${database_name}" \
    "DB_USER=${database_user}" \
    "DB_PASSWORD=${database_password}" \
    "DB_STATE_KEY=primary"
} >"${temporary_env}"
chmod 600 "${temporary_env}"
mv "${temporary_env}" "${env_file}"

echo "MariaDB and ${env_file} are configured. The generated password was not printed."
