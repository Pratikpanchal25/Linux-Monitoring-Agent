# Debian Packaging

This folder contains tooling to build a .deb package for Linux Monitoring Agent.

## Why this exists

End users should not need Go installed.
They can install Linux Monitoring Agent with a .deb package and manage it with systemd.

## Build the .deb (maintainer/release machine)

    chmod +x packaging/deb/build-deb.sh
    ./packaging/deb/build-deb.sh 1.0.0 amd64

Output package:

    dist/linux-monitoring-agent_1.0.0_amd64.deb

## Install the .deb (end user machine)

    sudo dpkg -i linux-monitoring-agent_1.0.0_amd64.deb
    sudo systemctl enable --now linux-monitoring-agent.service
    sudo systemctl status linux-monitoring-agent.service

Configuration file path:

    /etc/linux-monitoring-agent/config.yaml
