# Debian Packaging

This folder contains tooling to build a .deb package for watchd.

## Why this exists

End users should not need Go installed.
They can install watchd with a .deb package and manage it with systemd.

## Build the .deb (maintainer/release machine)

    chmod +x packaging/deb/build-deb.sh
    ./packaging/deb/build-deb.sh 1.0.0 amd64

Output package:

    dist/watchd_1.0.0_amd64.deb

## Install the .deb (end user machine)

    sudo dpkg -i watchd_1.0.0_amd64.deb
    sudo systemctl enable --now watchd.service
    sudo systemctl status watchd.service

Configuration file path:

    /etc/watchd/config.yaml
