# cpu-alert

cpu-alert is a lightweight Linux monitoring daemon.
It runs as a systemd service and sends email alerts for sustained high CPU or RAM usage.

## User Install (Only Steps You Need)

Download the `.deb` file from GitHub Releases and run only these commands:

```bash
sudo dpkg -i cpu-alert_v1.0.0_amd64.deb
sudo systemctl enable --now cpu-alert.service
sudo systemctl status cpu-alert.service
```

Then configure email and restart:

```bash
sudo nano /etc/cpu-alert/config.yaml
sudo systemctl restart cpu-alert.service
```

## Where To Download .deb

- Open: https://github.com/Pratikpanchal25/linux-cpu-alerts/releases/tag/v1.0.0
- Download asset: `cpu-alert_v1.0.0_amd64.deb`

## Config File Location

- `/etc/cpu-alert/config.yaml`

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
  from: "cpu-alert@example.com"
  smtp: "smtp.gmail.com:587"
  password: "your-app-password"
```

## Common Fix

If `.deb` install reports dependency errors:

```bash
sudo apt-get install -f -y
sudo dpkg -i cpu-alert_v1.0.0_amd64.deb
```

## Useful Service Commands

```bash
sudo systemctl restart cpu-alert.service
sudo systemctl stop cpu-alert.service
sudo systemctl start cpu-alert.service
sudo journalctl -u cpu-alert.service -f
```
