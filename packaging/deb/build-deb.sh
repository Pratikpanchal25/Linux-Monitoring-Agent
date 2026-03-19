#!/usr/bin/env sh
set -eu

# Build a Debian package for cpu-alert.
# This script is for maintainers/release builders, not end users.

APP_NAME="cpu-alert"
VERSION="${1:-1.0.0}"
ARCH="${2:-amd64}"
WORK_DIR="dist/${APP_NAME}_${VERSION}_${ARCH}"
DEBIAN_DIR="${WORK_DIR}/DEBIAN"

if ! command -v go >/dev/null 2>&1; then
  echo "Error: go is required to build the .deb package." >&2
  exit 1
fi

if ! command -v dpkg-deb >/dev/null 2>&1; then
  echo "Error: dpkg-deb is required to build the .deb package." >&2
  exit 1
fi

rm -rf "${WORK_DIR}"
mkdir -p "${DEBIAN_DIR}"
mkdir -p "${WORK_DIR}/usr/local/bin"
mkdir -p "${WORK_DIR}/etc/cpu-alert"
mkdir -p "${WORK_DIR}/lib/systemd/system"

# Build static single-binary target.
CGO_ENABLED=0 GOOS=linux GOARCH="${ARCH}" go build -trimpath -ldflags "-s -w" -o "${WORK_DIR}/usr/local/bin/${APP_NAME}" ./cmd/cpu-alert

# Install runtime files.
install -m 0644 ./configs/config.yaml "${WORK_DIR}/etc/cpu-alert/config.yaml"
install -m 0644 ./packaging/systemd/cpu-alert.service "${WORK_DIR}/lib/systemd/system/cpu-alert.service"

# Package metadata and lifecycle scripts.
cat >"${DEBIAN_DIR}/control" <<EOF
Package: ${APP_NAME}
Version: ${VERSION}
Section: admin
Priority: optional
Architecture: ${ARCH}
Maintainer: cpu-alert maintainers <maintainers@example.com>
Depends: systemd
Description: Lightweight CPU and memory alert daemon
 cpu-alert monitors CPU (/proc/stat) and memory (/proc/meminfo),
 then sends SMTP email alerts when usage is above configured thresholds
 for a sustained duration.
EOF

cat >"${DEBIAN_DIR}/conffiles" <<EOF
/etc/cpu-alert/config.yaml
EOF

cat >"${DEBIAN_DIR}/postinst" <<'EOF'
#!/usr/bin/env sh
set -eu

systemctl daemon-reload || true

echo "cpu-alert installed."
echo "Edit /etc/cpu-alert/config.yaml, then run:"
echo "  sudo systemctl enable --now cpu-alert.service"
EOF

cat >"${DEBIAN_DIR}/prerm" <<'EOF'
#!/usr/bin/env sh
set -eu

if [ "${1:-}" = "remove" ]; then
  systemctl disable --now cpu-alert.service || true
fi
EOF

cat >"${DEBIAN_DIR}/postrm" <<'EOF'
#!/usr/bin/env sh
set -eu

systemctl daemon-reload || true
EOF

chmod 0755 "${DEBIAN_DIR}/postinst" "${DEBIAN_DIR}/prerm" "${DEBIAN_DIR}/postrm"

dpkg-deb --build "${WORK_DIR}"

echo "Built package: ${WORK_DIR}.deb"
