#!/usr/bin/env bash
set -euo pipefail

readonly go_version="1.26.5"
readonly archive_name="go${go_version}.linux-amd64.tar.gz"
readonly archive_url="https://go.dev/dl/${archive_name}"
readonly expected_sha256="5c2c3b16caefa1d968a94c1daca04a7ca301a496d9b086e17ad77bb81393f053"
readonly install_root="/home/ubuntu/.local"
readonly go_root="${install_root}/go"

if [[ -x "${go_root}/bin/go" ]] && [[ "$("${go_root}/bin/go" version)" == *"go${go_version}"* ]]; then
  "${go_root}/bin/go" version
  exit 0
fi

install -d -m 755 "${install_root}" /home/ubuntu/.cache
temporary_dir=$(mktemp -d /home/ubuntu/.cache/relic-raid-go.XXXXXX)
archive_path="${temporary_dir}/${archive_name}"
trap 'rm -rf "${temporary_dir}"' EXIT

curl --fail --location --silent --show-error "${archive_url}" --output "${archive_path}"
printf '%s  %s\n' "${expected_sha256}" "${archive_path}" | sha256sum --check --status

if [[ -d "${go_root}" ]]; then
  mv "${go_root}" "${go_root}.bak.$(date +%Y%m%d-%H%M%S)"
fi
tar -C "${install_root}" -xzf "${archive_path}"

"${go_root}/bin/go" version
