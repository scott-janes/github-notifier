# Configuration

All settings are managed through the in-app Settings panel (gear icon in the receipt header). They are persisted to `~/.config/github-notifier/config.json`.

## Config File Reference

| Field | Type | Default | Description |
|---|---|---|---|
| `github_token` | string | `""` | GitHub classic PAT with `notifications` scope |
| `github_username` | string | `""` | Your GitHub username |
| `poll_interval_minutes` | int | `5` | How often to poll GitHub (1–60) |
| `schedule_days` | string[] | `["Mon","Tue","Wed","Thu"]` | Days polling is active |
| `schedule_start_hour` | int | `9` | Polling start (24h, 0–23) |
| `schedule_end_hour` | int | `17` | Polling end (24h, exclusive) |
| `sound_enabled` | bool | `true` | Play chime on new notifications |
| `auto_hide_seconds` | int | `10` | Seconds before the receipt auto-closes |
| `disable_drag` | bool | `false` | Lock dot position to prevent Cmd+drag |
| `dnd_enabled` | bool | `false` | Do Not Disturb active |
| `dnd_hours` | float | `2` | DND auto-expiry duration |
| `panel_opacity` | float | `0.95` | Receipt panel background opacity (0.3–1.0) |
| `window_x` / `window_y` | int | auto | Last window position |

## State File

Notification state is stored at `~/.config/github-notifier/state.json`:

```json
{
  "seen_ids": ["notification-id-1", "..."],
  "confirmed_ids": ["..."],
  "last_poll_time": "2026-05-15T12:00:00Z"
}
```

- `seen_ids` — all notifications that have been shown (prevents re-showing after restart)
- `confirmed_ids` — notifications you've confirmed
- `last_poll_time` — used as the `since` parameter for the next GitHub API poll

## GitHub Token

You need a [classic personal access token](https://github.com/settings/tokens) with the `notifications` scope. Fine-grained tokens do **not** support the Notifications API.

The token is stored securely in the **macOS Keychain** (via `security add-generic-password`), not in the config file. The `github_token` field in `config.json` is always written as empty — the real token lives in the Keychain under service `github-notifier-token`.

Existing tokens in `config.json` are automatically migrated to the Keychain on first load.
