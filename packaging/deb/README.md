# Debian Packaging

This folder contains tooling to build a .deb package for cpu-alert.

## Why this exists

End users should not need Go installed.
They can install cpu-alert with a .deb package and manage it with systemd.

## Build the .deb (maintainer/release machine)

    chmod +x packaging/deb/build-deb.sh
    ./packaging/deb/build-deb.sh 1.0.0 amd64

Output package:

    dist/cpu-alert_1.0.0_amd64.deb

## Install the .deb (end user machine)

    sudo dpkg -i cpu-alert_1.0.0_amd64.deb
    sudo systemctl enable --now cpu-alert.service
    sudo systemctl status cpu-alert.service

Configuration file path:

    /etc/cpu-alert/config.yaml
