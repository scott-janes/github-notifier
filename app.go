package main

import (
	"context"
	"log"
	"sort"

	"github.com/scott-janes/github-notifier/config"
	"github.com/scott-janes/github-notifier/gh"
	"github.com/scott-janes/github-notifier/mock"
	"github.com/scott-janes/github-notifier/poller"
	"github.com/scott-janes/github-notifier/state"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	cfg           *config.Config
	ghClient      *gh.Client
	pl            *poller.Poller
	mockPl        *mock.Poller
	st            *state.Store
	notifications []gh.Notification
	isMock        bool
}

func NewApp(isMock bool) *App {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("config load error: %v", err)
		cfg = config.Default()
	}

	cfg.MockMode = isMock

	return &App{
		cfg:    cfg,
		st:     state.New(),
		isMock: isMock,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	runtime.WindowSetBackgroundColour(ctx, 0, 0, 0, 0)
	runtime.WindowSetSize(ctx, 60, 60)
	a.positionWindow(ctx, 60)

	if a.isMock || a.cfg.GitHubToken != "" {
		a.startPoller()
	}
}

func (a *App) positionWindow(ctx context.Context, windowWidth int) {
	if a.cfg.WindowX != 0 && a.cfg.WindowY != 0 {
		runtime.WindowSetPosition(ctx, a.cfg.WindowX, a.cfg.WindowY)
		return
	}
	screens, err := runtime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		x := 1200 - (windowWidth - 60)
		runtime.WindowSetPosition(ctx, x, 20)
		return
	}
	screen := screens[0]
	x := screen.Width - windowWidth - 20
	y := 20
	runtime.WindowSetPosition(ctx, x, y)
}

func (a *App) startPoller() {
	if a.isMock {
		a.mockPl = mock.New(a.onNewNotifications)
		a.mockPl.Start()
		return
	}
	if a.ghClient == nil {
		a.ghClient = gh.NewClient(a.cfg.GitHubToken)
	}
	a.pl = poller.New(a.cfg, a.ghClient, a.st, a.onNewNotifications, a.onPollError)
	a.pl.Start()
}

func (a *App) onNewNotifications(notifs []gh.Notification) {
	a.notifications = append(a.notifications, notifs...)
	runtime.EventsEmit(a.ctx, "new-notifications", notifs)
}

func (a *App) onPollError(errMsg string) {
	runtime.EventsEmit(a.ctx, "api-error", errMsg)
}

func (a *App) shutdown(ctx context.Context) {
	if a.pl != nil {
		a.pl.Stop()
	}
	if a.mockPl != nil {
		a.mockPl.Stop()
	}
}

// --- Bound methods exposed to frontend ---

func (a *App) GetNotifications() []gh.Notification {
	if len(a.notifications) == 0 {
		return []gh.Notification{}
	}

	unconfirmed := make([]gh.Notification, 0, len(a.notifications))
	for _, n := range a.notifications {
		if !a.st.Confirmed(n.ID) {
			unconfirmed = append(unconfirmed, n)
		}
	}

	sort.Slice(unconfirmed, func(i, j int) bool {
		return unconfirmed[i].UpdatedAt.After(unconfirmed[j].UpdatedAt)
	})
	return unconfirmed
}

func (a *App) GetUnconfirmedCount() int {
	count := 0
	for _, n := range a.notifications {
		if !a.st.Confirmed(n.ID) {
			count++
		}
	}
	return count
}

func (a *App) ConfirmNotification(id string) {
	a.st.Confirm(id)
	a.st.Save()
	a.pruneConfirmed()
	runtime.EventsEmit(a.ctx, "count-updated", a.GetUnconfirmedCount())
}

func (a *App) ConfirmAll() {
	for _, n := range a.notifications {
		a.st.Confirm(n.ID)
	}
	a.st.Save()
	a.pruneConfirmed()
	runtime.EventsEmit(a.ctx, "count-updated", a.GetUnconfirmedCount())
}

func (a *App) pruneConfirmed() {
	filtered := make([]gh.Notification, 0, len(a.notifications))
	for _, n := range a.notifications {
		if !a.st.Confirmed(n.ID) {
			filtered = append(filtered, n)
		}
	}
	a.notifications = filtered
}

func (a *App) OpenInBrowser(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}

func (a *App) GetConfig() *config.Config {
	return a.cfg
}

func (a *App) SaveConfig(token, username string, interval int, days []string, startHour, endHour int, sound bool, soundPath string, autoHide int, disableDrag bool, dndEnabled bool, dndHours float64, panelOpacity float64) error {
	a.cfg.GitHubToken = token
	a.cfg.GitHubUsername = username
	a.cfg.PollIntervalMin = interval
	a.cfg.ScheduleDays = days
	a.cfg.ScheduleStartHour = startHour
	a.cfg.ScheduleEndHour = endHour
	a.cfg.SoundEnabled = sound
	a.cfg.SoundPath = soundPath
	a.cfg.AutoHideSeconds = autoHide
	a.cfg.DisableDrag = disableDrag
	a.cfg.DNDEnabled = dndEnabled
	a.cfg.DNDHours = dndHours
	a.cfg.PanelOpacity = panelOpacity

	if token != "" && !a.isMock {
		testClient := gh.NewClient(token)
		if err := testClient.ValidateToken(); err != nil {
			return err
		}
	}

	if err := a.cfg.Save(); err != nil {
		return err
	}

	// In mock mode, just restart the mock poller with new config
	if a.isMock {
		if a.mockPl != nil {
			a.mockPl.Stop()
		}
		a.mockPl = mock.New(a.onNewNotifications)
		a.mockPl.Start()
		return nil
	}

	if a.pl != nil {
		a.pl.Stop()
	}
	if a.mockPl != nil {
		a.mockPl.Stop()
	}
	a.ghClient = gh.NewClient(token)
	a.startPoller()
	return nil
}

func (a *App) ChooseSoundFile() (string, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose Notification Sound",
		Filters: []runtime.FileFilter{
			{DisplayName: "Audio Files", Pattern: "*.aiff;*.wav;*.mp3;*.m4a;*.caf"},
		},
	})
	if err != nil {
		return "", err
	}
	return file, nil
}

func (a *App) SaveWindowPosition(x, y int) {
	a.cfg.WindowX = x
	a.cfg.WindowY = y
	a.cfg.Save()
}

func (a *App) GetWindowPosition() (int, int) {
	return runtime.WindowGetPosition(a.ctx)
}

func (a *App) SetWindowPosition(x, y int) {
	runtime.WindowSetPosition(a.ctx, x, y)
}

func (a *App) ExpandWindow(height int) {
	x, y := runtime.WindowGetPosition(a.ctx)
	newX := x - 360
	if newX < 10 {
		newX = 10
	}
	runtime.WindowSetPosition(a.ctx, newX, y)
	runtime.WindowSetSize(a.ctx, 420, height)
}

func (a *App) ExpandToast() {
	x, y := runtime.WindowGetPosition(a.ctx)
	newX := x - 320
	if newX < 10 {
		newX = 10
	}
	runtime.WindowSetPosition(a.ctx, newX, y)
	runtime.WindowSetSize(a.ctx, 380, 220)
}

func (a *App) CollapseWindow() {
	x, y := runtime.WindowGetPosition(a.ctx)
	w, _ := runtime.WindowGetSize(a.ctx)
	dotX := x + w - 60
	if dotX < 0 {
		dotX = 0
	}
	dotY := y
	if dotY < 0 {
		dotY = 0
	}
	a.cfg.WindowX = dotX
	a.cfg.WindowY = dotY
	a.cfg.Save()
	runtime.WindowSetSize(a.ctx, 60, 60)
	runtime.WindowSetPosition(a.ctx, dotX, dotY)
}
