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
  formAutoHide: 10,
  formDisableDrag: false,
  formDNDEnabled: false,
  formDNDHours: 2,
  formPanelOpacity: 0.95,

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
      this.formAutoHide = cfg.auto_hide_seconds || 10
      this.formDisableDrag = cfg.disable_drag ?? false
      this.disableDrag = cfg.disable_drag ?? false
      this.formDNDEnabled = cfg.dnd_enabled ?? false
      this.formDNDHours = cfg.dnd_hours ?? 2
      this.formPanelOpacity = cfg.panel_opacity ?? 0.95
      if (cfg.dnd_enabled) {
        this._startDND(cfg.dnd_hours || 2)
      }
    }).catch(err => console.error('Failed to load config:', err))
  },

  // --- Grouped notifications ---

  groupedNotifications() {
    const groups = {}
    for (const n of this.notifications) {
      const num = this._prNumber(n.url)
      const key = num ? n.repo + '#' + num : n.id
      if (!groups[key]) {
        groups[key] = { ...n, count: 1, prNumber: num }
      } else {
        groups[key].count++
        groups[key].prNumber = num
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
    this.toastTimer = setTimeout(() => {
      if (!this.toastPaused) this.hideToast()
    }, 5000)
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
      this.formToken, this.formUsername, this.formInterval,
      this.formDays, this.formStartHour, this.formEndHour,
      this.formSound, this.formAutoHide, this.formDisableDrag,
      this.formDNDEnabled, this.formDNDHours, this.formPanelOpacity
    ).then(() => {
      this.config.github_token = this.formToken
      this.config.github_username = this.formUsername
      this.config.poll_interval_minutes = this.formInterval
      this.config.schedule_days = this.formDays
      this.config.schedule_start_hour = this.formStartHour
      this.config.schedule_end_hour = this.formEndHour
      this.config.sound_enabled = this.formSound
      this.config.auto_hide_seconds = this.formAutoHide
      this.config.disable_drag = this.formDisableDrag
      this.config.dnd_enabled = this.formDNDEnabled
      this.config.dnd_hours = this.formDNDHours
      this.config.panel_opacity = this.formPanelOpacity
      this.disableDrag = this.formDisableDrag

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
      review_requested: 'bg-yellow-500/20 text-yellow-400',
      mention: 'bg-purple-500/20 text-purple-400',
      comment: 'bg-blue-500/20 text-blue-400',
      author: 'bg-green-500/20 text-green-400',
      state_change: 'bg-orange-500/20 text-orange-400',
      subscribed: 'bg-gray-500/20 text-gray-400',
      team_mention: 'bg-pink-500/20 text-pink-400',
      approval_required: 'bg-red-500/20 text-red-400',
    }
    return classes[reason] || 'bg-gray-500/20 text-gray-400'
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

  playSound() {
    try {
      const ctx = new (window.AudioContext || window.webkitAudioContext)()
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
    } catch (e) {}
  },
}))

Alpine.start()
