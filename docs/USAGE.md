# Usage

## The Dot

The dot is the main interface — a small circle that sits on your desktop. Its appearance tells you the current state:

| State | Appearance |
|---|---|
| No notifications | Translucent white dot, outline bell icon |
| Unconfirmed notifications | Blue dot, filled bell icon, red badge with count |
| New notification just arrived | Blue dot pulsing for 20s |
| Do Not Disturb active | Indigo dot, moon icon |

| Action | Result |
|---|---|
| Click | Open/close the full receipt panel |
| Cmd+click+drag | Move the dot (hold Cmd, then click and drag) |

> Cmd+drag can be disabled in Settings → Lock position.

## Notification Toast

When new notifications arrive while the panel is closed, a receipt-printing toast appears below the dot:

- Each notification prints down one at a time (thermal printer effect)
- Tap **Confirm** to mark it read and hide
- Tap **Open** to view the PR in your browser
- Hover over the toast to pause the 5-second auto-hide timer
- Click the dot while a toast is showing to open the full receipt panel

If multiple notifications arrive in one poll, they queue up and play one after another.

## Full Receipt Panel

Click the dot to open the full panel. All unconfirmed notifications are listed, grouped by PR.

| Control | Location | Action |
|---|---|---|
| **Confirm** | Per notification | Mark one as read |
| **Open** | Per notification | Open PR in browser |
| **Confirm All** | Header, right | Mark all as read and close |
| **✕** | Header, left | Close panel without confirming |
| **⚙** | Header, left | Open settings |
| **🌙** | Header, left | Toggle Do Not Disturb |

The panel auto-hides after the configured duration (default 10s). Hover over it to pause.

## Do Not Disturb

When DND is active:
- Toasts are suppressed
- Sound is suppressed
- The dot shows a moon icon (indigo)
- Notifications still count up on the badge

Toggle DND from the receipt header (🌙 button) or configure auto-DND in Settings with a duration (30min, 1h, 2h, 4h, 8h). DND expires automatically after the set time.

## Settings

| Setting | Description |
|---|---|
| GitHub Token | Your PAT with `notifications` scope |
| GitHub Username | Your GitHub username |
| Poll Interval | Minutes between polls (1–60) |
| Active Days | Which days of the week to poll |
| Start/End Hour | Active polling window |
| Auto-hide | Seconds before the receipt closes |
| Sound | Enable/disable the notification chime |
| Lock Position | Disable Cmd+drag |
| Do Not Disturb | Auto-DND on launch + duration |
| Panel Opacity | Receipt transparency slider |

Click **Save & Restart Poller** to apply changes.
