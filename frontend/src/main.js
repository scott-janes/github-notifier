import './style.css'
import Alpine from 'alpinejs'

window.Alpine = Alpine

Alpine.data('app', () => ({
  expanded: false,
  collapsing: false,
  cmdHeld: false,
  notifications: [],
  unconfirmedCount: 0,
  settingsOpen: false,

  // Toast state
  toastVisible: false,
  toastAnimating: false,
  toastRetracting: false,
  toastNotifications: [],
  toastQueue: [],
  toastBusy: false,
  toastTimer: null,
  toastPaused: false,
  toastFooter: '········  ·  ·  ·  ········',
  _toastExpanded: false,

  // Pulse state
  pulseActive: false,
  _pulseTimer: null,

  // Polling state
  pollingActive: false,
  _pollingTimer: null,

  // DND state
  dndActive: false,
  _dndTimer: null,

  // API error
  apiError: '',

  config: {},
  disableDrag: false,
  formToken: '',
  formUsername: '',
  formInterval: 5,
  formDays: [],
  formStartHour: 9,
  formEndHour: 17,
  formSound: true,
  formSoundPath: '',
  formAutoHide: 10,
  formDisableDrag: false,
  formDNDEnabled: false,
  formDNDHours: 2,
  formPanelOpacity: 0.95,
  formTheme: 'default',
  formCustomCSS: '',
  _themeStyleEl: null,

  weekDays: ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'],

  init() {
    document.addEventListener('keydown', (e) => {
      if (e.metaKey || e.ctrlKey) this.cmdHeld = true
    })
    document.addEventListener('keyup', (e) => {
      if (!e.metaKey && !e.ctrlKey) this.cmdHeld = false
    })
    window.addEventListener('blur', () => { this.cmdHeld = false })

    this.loadData()

    this._themeStyleEl = document.createElement('style')
    this._themeStyleEl.id = 'custom-theme'
    document.head.appendChild(this._themeStyleEl)

    window.runtime.EventsOn('new-notifications', (notifs) => {
      if (!notifs || notifs.length === 0) return
      this.apiError = ''
      this.notifications = [...notifs, ...this.notifications]
      this.unconfirmedCount += notifs.length
      this.settingsOpen = false

      this.pulseActive = true
      if (this._pulseTimer) clearTimeout(this._pulseTimer)
      this._pulseTimer = setTimeout(() => { this.pulseActive = false }, 20000)

      if (this.config.sound_enabled && !this.dndActive) {
        this.playSound()
      }

      if (this.expanded) return
      // DND suppresses toasts
      if (this.dndActive) return

      for (const n of notifs) {
        this.toastQueue.push(n)
      }
      this.processToastQueue()
    })

    window.runtime.EventsOn('count-updated', (count) => {
      this.unconfirmedCount = count
    })

    window.runtime.EventsOn('api-error', (msg) => {
      this.apiError = msg || 'GitHub API error'
      setTimeout(() => { this.apiError = '' }, 8000)
    })

    window.runtime.EventsOn('poll-complete', () => {
      this.pollingActive = true
      if (this._pollingTimer) clearTimeout(this._pollingTimer)
      this._pollingTimer = setTimeout(() => { this.pollingActive = false }, 800)
    })
  },

  loadData() {
    window.go.main.App.GetNotifications().then(notifs => {
      this.notifications = notifs || []
    }).catch(err => console.error('Failed to load notifications:', err))

    window.go.main.App.GetUnconfirmedCount().then(count => {
      this.unconfirmedCount = count || 0
    }).catch(err => console.error('Failed to load count:', err))

    window.go.main.App.GetConfig().then(cfg => {
      if (!cfg) return
      this.config = cfg
      this.formToken = cfg.github_token || ''
      this.formUsername = cfg.github_username || ''
      this.formInterval = cfg.poll_interval_minutes || 5
      this.formDays = cfg.schedule_days || []
      this.formStartHour = cfg.schedule_start_hour ?? 9
      this.formEndHour = cfg.schedule_end_hour ?? 17
      this.formSound = cfg.sound_enabled ?? true
      this.formSoundPath = cfg.sound_path || ''
      this.formAutoHide = cfg.auto_hide_seconds || 10
      this.formDisableDrag = cfg.disable_drag ?? false
      this.disableDrag = cfg.disable_drag ?? false
      this.formDNDEnabled = cfg.dnd_enabled ?? false
      this.formDNDHours = cfg.dnd_hours ?? 2
      this.formPanelOpacity = cfg.panel_opacity ?? 0.95
      this.formTheme = cfg.theme || 'default'
      this.formCustomCSS = cfg.custom_theme_css || ''
      this.applyTheme()
      if (cfg.dnd_enabled) {
        this._startDND(cfg.dnd_hours || 2)
      }
    }).catch(err => console.error('Failed to load config:', err))
  },

  applyTheme() {
    const theme = this.formTheme || 'default'
    document.documentElement.dataset.theme = theme === 'default' ? '' : theme
    if (this._themeStyleEl) {
      this._themeStyleEl.textContent = theme === 'custom' ? (this.formCustomCSS || '') : ''
    }
  },

  // --- Grouped notifications ---

  groupedNotifications() {
    const groups = {}
    for (const n of this.notifications) {
      const num = this._prNumber(n.url)
      const key = num ? n.repo + '#' + num : n.id
      if (!groups[key]) {
        groups[key] = { ...n, count: 1, prNumber: num, reasons: [n.reason] }
      } else {
        groups[key].count++
        if (!groups[key].reasons.includes(n.reason)) {
          groups[key].reasons.push(n.reason)
        }
        groups[key].updated_at = n.updated_at
      }
    }
    return Object.values(groups)
  },

  _prNumber(url) {
    if (!url) return null
    const m = url.match(/\/pull\/(\d+)$/)
    return m ? m[1] : null
  },

  groupLabel(n) {
    if (n.count > 1) {
      return n.title + ' (+' + (n.count - 1) + ' more)'
    }
    return n.title
  },

  // --- DND ---

  toggleDND() {
    if (this.dndActive) {
      this._stopDND()
    } else {
      this._startDND(this.config.dnd_hours || 2)
    }
  },

  _startDND(hours) {
    this.dndActive = true
    if (this._dndTimer) clearTimeout(this._dndTimer)
    this._dndTimer = setTimeout(() => {
      this.dndActive = false
    }, hours * 3600 * 1000)
  },

  _stopDND() {
    this.dndActive = false
    if (this._dndTimer) {
      clearTimeout(this._dndTimer)
      this._dndTimer = null
    }
  },

  dndRemaining() {
    // approximated: shown as "DND" label, we don't track the exact end
    return ''
  },

  // --- Toast ---

  processToastQueue() {
    if (this.toastBusy || this.toastQueue.length === 0) return

    const next = this.toastQueue.shift()
    this.toastBusy = true
    this.toastNotifications = [next]
    this.toastAnimating = false
    this.toastRetracting = false
    this.toastPaused = false
    this.toastVisible = false

    if (!this._toastExpanded) {
      window.go.main.App.ExpandToast()
      this._toastExpanded = true
    }

    setTimeout(() => {
      this.toastVisible = true
      this.toastAnimating = true
    }, 80)
    this.startToastTimer()
  },

  hideToast() {
    this.toastAnimating = false
    this.toastRetracting = true
    this.cancelToastTimer()

    setTimeout(() => {
      this.toastVisible = false
      this.toastRetracting = false
      this.toastBusy = false

      if (this.toastQueue.length > 0) {
        setTimeout(() => {
          this.toastNotifications = []
          this.processToastQueue()
        }, 400)
      } else {
        this.toastNotifications = []
        this.toastQueue = []
        this._toastExpanded = false
        window.go.main.App.CollapseWindow()
      }
    }, 320)
  },

  toastConfirm(id) {
    window.go.main.App.ConfirmNotification(id)
    this.notifications = this.notifications.filter(n => n.id !== id)
    this.unconfirmedCount = Math.max(0, this.unconfirmedCount - 1)
    this.toastQueue = []
    this.hideToast()
  },

  toastOpen(url) {
    window.go.main.App.OpenInBrowser(url)
  },

  startToastTimer() {
    this.cancelToastTimer()
    const delay = (this.config.auto_hide_seconds || 5) * 1000
    this.toastTimer = setTimeout(() => {
      if (!this.toastPaused) this.hideToast()
    }, delay)
  },

  cancelToastTimer() {
    if (this.toastTimer) { clearTimeout(this.toastTimer); this.toastTimer = null }
  },

  pauseToastTimer() {
    this.toastPaused = true; this.cancelToastTimer()
  },

  resumeToastTimer() {
    this.toastPaused = false; this.startToastTimer()
  },

  dismissToast() {
    this.toastQueue = []; this.hideToast()
  },

  confirm(id) {
    window.go.main.App.ConfirmNotification(id)
    this.notifications = this.notifications.filter(n => n.id !== id)
    this.unconfirmedCount = Math.max(0, this.unconfirmedCount - 1)
  },

  confirmAll() {
    window.go.main.App.ConfirmAll()
    this.notifications = []
    this.unconfirmedCount = 0
    this.collapse()
  },

  closePanel() { this.collapse() },

  openBrowser(url) { window.go.main.App.OpenInBrowser(url) },

  toggleExpand() {
    if (this.expanded) {
      this.collapse()
    } else {
      this.toastQueue = []
      this.toastBusy = false
      this.toastVisible = false
      this.toastAnimating = false
      this.toastRetracting = false
      this.cancelToastTimer()
      this.toastNotifications = []
      const wasToast = this._toastExpanded
      this._toastExpanded = false

      if (wasToast) {
        window.go.main.App.CollapseWindow().then(() => { this.expand() })
      } else {
        this.expand()
      }
    }
  },

  handleDotMousedown(event) {
    this._dotStartX = event.screenX; this._dotStartY = event.screenY
  },

  handleDotMouseup(event) {
    if (this.cmdHeld) return
    const dx = Math.abs(event.screenX - this._dotStartX)
    const dy = Math.abs(event.screenY - this._dotStartY)
    if (dx < 10 && dy < 10) this.toggleExpand()
  },

  expand() {
    this.expanded = true
    window.go.main.App.ExpandWindow(500)
    this.startAutoHide()
  },

  collapse() {
    this.settingsOpen = false
    this.cancelAutoHide()
    this.expanded = false
    setTimeout(() => { window.go.main.App.CollapseWindow() }, 50)
  },

  autoHideTimer: null,

  startAutoHide() {
    this.cancelAutoHide()
    const delay = (this.config.auto_hide_seconds || 10) * 1000
    this.autoHideTimer = setTimeout(() => { this.collapse() }, delay)
  },

  cancelAutoHide() {
    if (this.autoHideTimer) { clearTimeout(this.autoHideTimer); this.autoHideTimer = null }
  },

  saveSettings() {
    window.go.main.App.SaveConfig(
      this.formToken, this.formUsername, parseInt(this.formInterval) || 5,
      this.formDays, parseInt(this.formStartHour) || 9, parseInt(this.formEndHour) || 17,
      this.formSound, this.formSoundPath, parseInt(this.formAutoHide) || 10,
      this.formDisableDrag, this.formDNDEnabled, parseFloat(this.formDNDHours) || 2,
      parseFloat(this.formPanelOpacity) || 0.95, this.formTheme, this.formCustomCSS
    ).then(() => {
      this.config.github_token = this.formToken
      this.config.github_username = this.formUsername
      this.config.poll_interval_minutes = this.formInterval
      this.config.schedule_days = this.formDays
      this.config.schedule_start_hour = this.formStartHour
      this.config.schedule_end_hour = this.formEndHour
      this.config.sound_enabled = this.formSound
      this.config.sound_path = this.formSoundPath
      this.config.auto_hide_seconds = this.formAutoHide
      this.config.disable_drag = this.formDisableDrag
      this.config.dnd_enabled = this.formDNDEnabled
      this.config.dnd_hours = this.formDNDHours
      this.config.panel_opacity = this.formPanelOpacity
      this.config.theme = this.formTheme
      this.config.custom_theme_css = this.formCustomCSS
      this.disableDrag = this.formDisableDrag
      this.applyTheme()

      if (this.formDNDEnabled) {
        this._startDND(this.formDNDHours)
      } else {
        this._stopDND()
      }

      this.settingsOpen = false
      this.collapse()
    }).catch(err => {
      this.apiError = 'Settings save failed: ' + (err || 'check token')
      setTimeout(() => { this.apiError = '' }, 8000)
    })
  },

  chooseSoundFile() {
    window.go.main.App.ChooseSoundFile().then(path => {
      if (path) this.formSoundPath = path
    }).catch(err => console.error('Failed to choose sound:', err))
  },

  resetPosition() {
    window.go.main.App.ResetWindowPosition().catch(err => console.error('Failed to reset position:', err))
  },

  toggleDay(day) {
    const idx = this.formDays.indexOf(day)
    if (idx >= 0) {
      this.formDays = this.formDays.filter(d => d !== day)
    } else {
      this.formDays = [...this.formDays, day]
    }
  },

  reasonLabel(reason) {
    const labels = {
      review_requested: 'Review Requested',
      mention: 'Mentioned',
      comment: 'New Comment',
      author: 'Your PR Updated',
      state_change: 'State Changed',
      subscribed: 'New Activity',
      team_mention: 'Team Mention',
      approval_required: 'Approval Required',
    }
    return labels[reason] || reason
  },

  reasonClass(reason) {
    const classes = {
      review_requested: 'theme-badge-review',
      mention: 'theme-badge-mention',
      comment: 'theme-badge-comment',
      author: 'theme-badge-author',
      state_change: 'theme-badge-state',
      subscribed: 'theme-badge-subscribed',
      team_mention: 'theme-badge-team',
      approval_required: 'theme-badge-approval',
    }
    return classes[reason] || 'theme-badge-subscribed'
  },

  timeAgo(t) {
    if (!t) return ''
    const diff = Date.now() - new Date(t).getTime()
    const mins = Math.floor(diff / 60000)
    if (mins < 1) return 'just now'
    if (mins < 60) return mins + 'm ago'
    const hours = Math.floor(mins / 60)
    if (hours < 24) return hours + 'h ago'
    const days = Math.floor(hours / 24)
    if (days < 7) return days + 'd ago'
    return new Date(t).toLocaleDateString()
  },

  _audioCtx: null,

  playSound() {
    try {
      if (!this._audioCtx) {
        this._audioCtx = new (window.AudioContext || window.webkitAudioContext)()
      }
      const ctx = this._audioCtx
      if (ctx.state === 'suspended') {
        ctx.resume()
      }
      const o1 = ctx.createOscillator(); const g1 = ctx.createGain()
      o1.connect(g1); g1.connect(ctx.destination)
      o1.frequency.value = 523.25; o1.type = 'sine'
      g1.gain.setValueAtTime(0, ctx.currentTime)
      g1.gain.linearRampToValueAtTime(0.1, ctx.currentTime + 0.02)
      g1.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.12)
      o1.start(ctx.currentTime); o1.stop(ctx.currentTime + 0.12)

      const o2 = ctx.createOscillator(); const g2 = ctx.createGain()
      o2.connect(g2); g2.connect(ctx.destination)
      o2.frequency.value = 659.25; o2.type = 'sine'
      g2.gain.setValueAtTime(0, ctx.currentTime + 0.08)
      g2.gain.linearRampToValueAtTime(0.1, ctx.currentTime + 0.1)
      g2.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.3)
      o2.start(ctx.currentTime + 0.08); o2.stop(ctx.currentTime + 0.3)
    } catch (e) {
      console.warn('playSound failed:', e)
    }
  },
}))

Alpine.start()
