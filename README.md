# portwatch

Lightweight CLI daemon that monitors open ports and alerts on unexpected changes via webhook or desktop notification.

---

## Installation

```bash
go install github.com/yourusername/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/portwatch.git && cd portwatch && go build -o portwatch .
```

---

## Usage

Start the daemon with a config file:

```bash
portwatch --config config.yaml
```

Example `config.yaml`:

```yaml
interval: 30s
baseline:
  - 22
  - 80
  - 443
alerts:
  webhook: "https://hooks.example.com/notify"
  desktop: true
```

portwatch will scan open ports on the specified interval and send an alert whenever a port opens or closes outside of your defined baseline.

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `config.yaml` | Path to config file |
| `--interval` | `30s` | Polling interval |
| `--once` | `false` | Run a single scan and exit |

```bash
# Run a one-time scan and print results
portwatch --once
```

---

## How It Works

1. Captures a snapshot of currently open ports
2. Compares against the defined baseline and previous snapshot
3. Fires a webhook POST or desktop notification on any unexpected change

---

## License

MIT © 2024 yourusername