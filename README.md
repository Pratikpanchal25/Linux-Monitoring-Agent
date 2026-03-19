# Linux Monitoring Agent

Linux Monitoring Agent is a lightweight Linux monitoring daemon.
It runs as a systemd service and sends email alerts for sustained high CPU or RAM usage.

## User Install (Only Steps You Need)

Download the `.deb` file from GitHub Releases and run only these commands:

```bash
sudo dpkg -i linux-monitoring-agent_1.0.0_amd64.deb
sudo systemctl enable --now linux-monitoring-agent.service
sudo systemctl status linux-monitoring-agent.service
```

Then configure email and restart:

```bash
sudo nano /etc/linux-monitoring-agent/config.yaml
sudo systemctl restart linux-monitoring-agent.service
```

## Where To Download .deb

- Open: https://github.com/Pratikpanchal25/linux-cpu-alerts/releases/latest
- Download asset: `linux-monitoring-agent_1.0.0_amd64.deb`

## Config File Location

- `/etc/linux-monitoring-agent/config.yaml`

Example config:

```yaml
thresholds:
  cpu: 80
  memory: 75  
interval: 10
duration: 120
cooldown: 300
email:
  to: "your-email@example.com"
  from: "linux-monitoring-agent@example.com"
  smtp: "smtp.gmail.com:587"
  password: "your-app-password"
```

## Common Fix

If `.deb` install reports dependency errors:

```bash
sudo apt-get install -f -y
sudo dpkg -i linux-monitoring-agent_1.0.0_amd64.deb
```

## Useful Service Commands

```bash
sudo systemctl restart linux-monitoring-agent.service
sudo systemctl stop linux-monitoring-agent.service
sudo systemctl start linux-monitoring-agent.service
sudo journalctl -u linux-monitoring-agent.service -f
```
