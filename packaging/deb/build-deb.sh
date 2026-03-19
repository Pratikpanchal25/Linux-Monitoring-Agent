#!/usr/bin/env sh
set -eu

# Build a Debian package for Linux Monitoring Agent.
# This script is for maintainers/release builders, not end users.

APP_NAME="linux-monitoring-agent"
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
mkdir -p "${WORK_DIR}/etc/linux-monitoring-agent"
mkdir -p "${WORK_DIR}/lib/systemd/system"

# Build static single-binary target.
CGO_ENABLED=0 GOOS=linux GOARCH="${ARCH}" go build -trimpath -ldflags "-s -w" -o "${WORK_DIR}/usr/local/bin/${APP_NAME}" ./cmd/linux-monitoring-agent

# Install runtime files.
install -m 0644 ./configs/config.yaml "${WORK_DIR}/etc/linux-monitoring-agent/config.yaml"
install -m 0644 ./packaging/systemd/linux-monitoring-agent.service "${WORK_DIR}/lib/systemd/system/linux-monitoring-agent.service"

# Package metadata and lifecycle scripts.
cat >"${DEBIAN_DIR}/control" <<EOF
Package: ${APP_NAME}
Version: ${VERSION}
Section: admin
Priority: optional
Architecture: ${ARCH}
Maintainer: Linux Monitoring Agent maintainers <maintainers@example.com>
Depends: systemd
Description: Linux Monitoring Agent for CPU and memory alerts
 Linux Monitoring Agent monitors CPU (/proc/stat) and memory (/proc/meminfo),
 then sends SMTP email alerts when usage is above configured thresholds
 for a sustained duration.
EOF

cat >"${DEBIAN_DIR}/conffiles" <<EOF
/etc/linux-monitoring-agent/config.yaml
EOF

cat >"${DEBIAN_DIR}/postinst" <<'EOF'
#!/usr/bin/env sh
set -eu

systemctl daemon-reload || true

echo "linux-monitoring-agent installed."
echo "Edit /etc/linux-monitoring-agent/config.yaml, then run:"
echo "  sudo systemctl enable --now linux-monitoring-agent.service"
EOF

cat >"${DEBIAN_DIR}/prerm" <<'EOF'
#!/usr/bin/env sh
set -eu

if [ "${1:-}" = "remove" ]; then
  systemctl disable --now linux-monitoring-agent.service || true
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
