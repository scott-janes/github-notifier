# Architecture

## Overview

The app is built with [Wails v2](https://wails.io), which wraps a Go backend in a native macOS WebView. The frontend communicates with Go via Wails bindings (direct method calls) and events (push notifications).

```
┌──────────────────────────────────────────────┐
│  Wails v2 (Go + WKWebView)                   │
│  ┌──────────────┐   ┌─────────────────────┐  │
│  │ Go Backend   │   │ Frontend (WebView)  │  │
│  │              │   │                     │  │
│  │ main.go      │   │ Alpine.js           │  │
│  │ app.go       │   │ Tailwind CSS        │  │
│  │              │   │ index.html          │  │
│  │ config/      │◄──┤                     │  │
│  │ gh/          │──►│ Wails bindings      │  │
│  │ poller/      │   │ (window.go.main.*)  │  │
│  │ state/       │   │                     │  │
│  │ sound/       │   │ Wails events        │  │
│  │ mock/        │──►│ (EventsOn/EventsEmit)│  │
│  └──────────────┘   └─────────────────────┘  │
└──────────────────────────────────────────────┘
```

## Backend Packages

### `main.go`
Wails application entry point. Configures the window (frameless, always-on-top, transparent, 60×60 initial size) and binds the App struct.

### `app.go`
The `App` struct exposes all methods callable from the frontend:

- **Data**: `GetNotifications()`, `GetUnconfirmedCount()`, `GetConfig()`
- **Actions**: `ConfirmNotification()`, `ConfirmAll()`, `OpenInBrowser()`
- **Window**: `ExpandWindow()`, `ExpandToast()`, `CollapseWindow()`, `GetWindowPosition()`, `SetWindowPosition()`
- **Config**: `SaveConfig()`
- **Events**: Emits `new-notifications`, `count-updated`, `api-error`

### `config/`
Loads and saves `~/.config/github-notifications/config.json`. Default values are returned if the file doesn't exist.

### `gh/`
GitHub REST API v3 client. Fetches from `/notifications` with pagination support (follows `Link: rel="next"` headers, up to 100 per page). Filters to `PullRequest` type only. Converts API URLs to browser URLs.

### `poller/`
A background goroutine that polls GitHub on a configurable interval. Checks the schedule (day of week + hour range) before polling. Compares returned notifications with the store's `seenIDs` to find new ones. Emits errors via the `ErrorHandler` callback.

### `state/`
Thread-safe persistence of notification IDs and last poll time. File is read on startup and written after each poll or confirm action. Uses `sync.Mutex` for concurrent access safety.

### `sound/`
Plays macOS system sounds via `afplay`. Uses `/System/Library/Sounds/Ping.aiff`.

### `mock/`
Generates fake notifications for testing. Includes 14 realistic-looking PRs with various reasons (review_requested, mention, comment, etc.). Emits 1–3 notifications every 15 seconds.

## Frontend

Built with Alpine.js (reactive state) and Tailwind CSS (utility-first styling), bundled with Vite.

### Key Alpine States

| State | Purpose |
|---|---|
| `expanded` | Full receipt panel visible |
| `toastVisible` | Notification toast visible |
| `toastQueue` | Queue of notifications waiting to display |
| `dndActive` | Do Not Disturb active |
| `pulseActive` | Dot pulsing animation (20s timer) |
| `apiError` | Current error message to display |
| `cmdHeld` | Cmd key tracking for drag |

## Data Flow

```
GitHub API
    │
    ▼
poller.go: poll() → gh.GetNotifications()
    │
    ├─ error → ErrorHandler → runtime.EventsEmit("api-error")
    │
    └─ new notifs found
         │
         ▼
    app.go: onNewNotifications()
         │
         ▼
    runtime.EventsEmit("new-notifications", notifs)
         │
         ▼
    main.js: EventsOn("new-notifications")
         │
         ├─ add to this.notifications[]
         ├─ increment unconfirmedCount
         ├─ if collapsed: queue for toast
         └─ if expanded: update panel inline
```

## Window States

| State | Size | Visible |
|---|---|---|
| Collapsed (dot) | 60×60 | Only the dot |
| Toast printing | 380×220 | Dot + toast |
| Full receipt | 420×500 | Full panel |
