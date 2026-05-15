<p align="center">
  <img src="build/appicon.png" width="96" height="96" alt="">
</p>

<h1 align="center">GitHub Notifier</h1>

<p align="center">
  A floating dot for your macOS desktop that polls GitHub for PR notifications and prints them like a receipt.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/platform-macOS-blue?style=flat-square">
  <img src="https://img.shields.io/badge/go-%3E%3D1.23-blue?style=flat-square">
  <img src="https://img.shields.io/badge/wails-v2.12-blue?style=flat-square">
  <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square">
</p>

---

## Quick Start

```bash
open github-notifier.app --args --mock    # try it with fake data
open github-notifier.app                   # real GitHub notifications
```

See [Installation](docs/INSTALL.md) for building from source and [Configuration](docs/CONFIGURATION.md) for setup.

## Features at a Glance

- Minimal floating dot — always-on-top, drag anywhere with Cmd+click
- Receipt-printing notification toasts — each prints down one at a time
- Full receipt panel — grouped by PR, with Confirm/Open/DND/Settings
- Scheduled polling — configure days, hours, and Do Not Disturb
- Persistent state — confirmed notifications stay across restarts

## Docs

| Document | What's in it |
|---|---|
| [Installation](docs/INSTALL.md) | Prerequisites, building from source, dev setup |
| [Configuration](docs/CONFIGURATION.md) | All settings, config file reference, state |
| [Usage](docs/USAGE.md) | Dot controls, receipt panel, toasts, DND |
| [Architecture](docs/ARCHITECTURE.md) | Project structure, data flow, backend packages |

## Build

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
git clone https://github.com/scott-janes/github-notifier
cd github-notifier && wails build
open build/bin/github-notifier.app
```

## License

MIT
