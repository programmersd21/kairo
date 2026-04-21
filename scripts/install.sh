#!/usr/bin/env bash
set -euo pipefail

OWNER="programmersd21"
REPO="kairo"
APP="kairo"

say() { printf '%s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

need_cmd() { command -v "$1" >/dev/null 2>&1 || die "missing dependency: $1"; }

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$os" in
  linux) os="linux" ;;
  darwin) os="darwin" ;;
  *) die "unsupported OS: $(uname -s)" ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="x86_64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) die "unsupported architecture: $arch" ;;
esac

ext="tar.gz"
asset="${APP}_${os}_${arch}.${ext}"
base="https://github.com/${OWNER}/${REPO}/releases/latest/download"
archive_url="${base}/${asset}"
checksums_url="${base}/checksums.txt"

tmp="$(mktemp -d)"
cleanup() { rm -rf "$tmp"; }
trap cleanup EXIT

need_cmd curl
need_cmd tar

install_dir="${HOME}/.local/bin"
if [[ -z "${HOME:-}" || ! -d "${HOME}" ]]; then
  install_dir="/usr/local/bin"
fi

if ! mkdir -p "$install_dir" 2>/dev/null; then
  if [[ "$install_dir" != "/usr/local/bin" ]]; then
    install_dir="/usr/local/bin"
    mkdir -p "$install_dir" 2>/dev/null || die "cannot create install dir (try: sudo mkdir -p /usr/local/bin)"
  else
    die "cannot create install dir (try: sudo mkdir -p /usr/local/bin)"
  fi
fi

checksums_path="${tmp}/checksums.txt"
archive_path="${tmp}/${asset}"

say "Downloading ${APP} (${os}/${arch})..."
curl -fsSL "$checksums_url" -o "$checksums_path"
curl -fsSL "$archive_url" -o "$archive_path"

sum_tool=""
if command -v sha256sum >/dev/null 2>&1; then
  sum_tool="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  sum_tool="shasum -a 256"
else
  die "missing dependency: sha256sum (or shasum)"
fi

want_sum="$(awk -v file="$asset" '$2==file || $2=="*"file { print $1; exit }' "$checksums_path")"
[[ -n "$want_sum" ]] || die "checksum for ${asset} not found in checksums.txt"

got_sum="$(eval "$sum_tool \"${archive_path}\"" | awk '{print $1}')"
[[ "${got_sum}" == "${want_sum}" ]] || die "checksum mismatch for ${asset}"

tar -xzf "$archive_path" -C "$tmp"
[[ -f "${tmp}/${APP}" ]] || die "archive did not contain ${APP}"

chmod +x "${tmp}/${APP}"
cp "${tmp}/${APP}" "${install_dir}/${APP}"

say "Installed to ${install_dir}/${APP}"

if [[ ":$PATH:" != *":${install_dir}:"* ]]; then
  if [[ "$install_dir" == "${HOME}/.local/bin" ]]; then
    if [[ -n "${SHELL:-}" && "${SHELL}" == *"zsh"* ]]; then
      profile="${HOME}/.zprofile"
    elif [[ -n "${SHELL:-}" && "${SHELL}" == *"bash"* ]]; then
      profile="${HOME}/.bashrc"
    else
      profile="${HOME}/.profile"
    fi

    if [[ -w "$(dirname "$profile")" ]]; then
      touch "$profile" 2>/dev/null || true
    fi

    if [[ -w "$profile" ]] && ! grep -qE '(^|\s)export PATH=.*\.local/bin' "$profile" 2>/dev/null; then
      {
        printf '\n# Added by kairo installer\n'
        printf 'export PATH=\"$HOME/.local/bin:$PATH\"\n'
      } >>"$profile"
      say "Added ${HOME}/.local/bin to PATH via ${profile}"
    fi
  fi

  say ""
  say "PATH update required:"
  say "  Run: export PATH=\"${install_dir}:$PATH\""
  if [[ "$install_dir" == "${HOME}/.local/bin" ]]; then
    say "  Then restart your shell (or source your profile) to persist it."
  fi
fi

say ""
say "Verify:"
say "  ${APP} version"

