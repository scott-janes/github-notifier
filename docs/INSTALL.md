# Installation

## Prerequisites

- macOS (for `afplay` system sound)
- [Go 1.23+](https://go.dev/dl/)
- [Node.js](https://nodejs.org/) and npm
- A [GitHub personal access token](https://github.com/settings/tokens) (classic) with the **notifications** scope

## Build from Source

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone and build
git clone https://github.com/scottjanes/github-notifier.git
cd github-notifier
wails build

# Launch
open build/bin/github-notifier.app
```

## Development (Hot Reload)

```bash
wails dev
```

This starts a Vite dev server for the frontend with hot-reload. Go file changes auto-rebuild.

## Mock Mode (No Token Needed)

```bash
open build/bin/github-notifier.app --args --mock
```

Or from the command line:

```bash
build/bin/github-notifier.app/Contents/MacOS/github-notifier --mock
```

Mock mode generates 1–3 fake PR notifications every 15 seconds so you can test the full UI without connecting to GitHub.
