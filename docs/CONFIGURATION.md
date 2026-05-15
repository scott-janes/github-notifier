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
| `sound_path` | string | `/System/Library/Sounds/Ping.aiff` | Custom notification sound file path |
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

### Creating a Token

1. Go to [GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)](https://github.com/settings/tokens)
2. Click **Generate new token (classic)**
3. Give it a name (e.g. "GitHub Notifier")
4. Under **Scopes**, select only **`notifications`**
5. Click **Generate token**
6. Copy the generated token (it looks like `ghp_xxxxxxxxxxxxxxxxxxxx`)
7. Open GitHub Notifier, click the gear icon (⚙) in the receipt header, paste the token into the **GitHub Token** field, and click **Save & Restart Poller**

> The token is stored securely in the **macOS Keychain** (via `security add-generic-password`), not in the config file. The `github_token` field in `config.json` is always written as empty — the real token lives in the Keychain under service `github-notifier-token`. Existing tokens in `config.json` are automatically migrated to the Keychain on first load.

### What You Get Notified Of

The app polls the [GitHub Notifications API](https://docs.github.com/en/rest/activity/notifications) for **PullRequests only** (issues are filtered out). It picks up notifications where you are **directly involved**:

| Reason | When it fires |
|---|---|
| `review_requested` | Someone requested your review on a PR |
| `mention` | You were @mentioned in a PR |
| `author` | Someone commented on or reviewed your PR |
| `comment` | You commented on a PR and someone else also commented |
| `state_change` | A PR you're involved in was merged, closed, or reopened |
| `assign` | Someone assigned you to a PR |
| `subscribed` | You're watching the repo and there's activity |

> You will **not** be notified for your own actions (self-assigning, your own comments). GitHub's API does not generate notifications for self-directed activity.
