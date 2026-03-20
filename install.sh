#!/usr/bin/env sh
set -eu

APP_NAME="watchd"
BIN_PATH="/usr/local/bin/${APP_NAME}"
CONF_DIR="/etc/${APP_NAME}"
SERVICE_PATH="/etc/systemd/system/${APP_NAME}.service"

echo "Building ${APP_NAME}..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o "${APP_NAME}" ./cmd/watchd

echo "Installing binary to ${BIN_PATH}..."
sudo install -m 0755 "${APP_NAME}" "${BIN_PATH}"

echo "Creating ${CONF_DIR}..."
sudo mkdir -p "${CONF_DIR}"

if [ ! -f "${CONF_DIR}/config.yaml" ]; then
  echo "Installing default config to ${CONF_DIR}/config.yaml..."
  sudo install -m 0600 ./configs/config.yaml "${CONF_DIR}/config.yaml"
else
  echo "Config already exists at ${CONF_DIR}/config.yaml, leaving it unchanged."
fi

echo "Installing systemd service..."
sudo install -m 0644 ./packaging/systemd/watchd.service "${SERVICE_PATH}"

echo "Reloading systemd and enabling service..."
sudo systemctl daemon-reload
sudo systemctl enable --now "${APP_NAME}.service"

echo "Done. Check status with: sudo systemctl status ${APP_NAME}.service"
