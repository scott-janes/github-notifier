package poller

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/scott-janes/github-notifier/config"
	"github.com/scott-janes/github-notifier/gh"
	"github.com/scott-janes/github-notifier/state"
	"github.com/scott-janes/github-notifier/sound"
)

const initialPollDelay = 2 * time.Second

type NotifHandler func([]gh.Notification)
type ErrorHandler func(string)

type Poller struct {
	cfg          *config.Config
	gh           *gh.Client
	state        *state.Store
	handler      NotifHandler
	errHandler   ErrorHandler
	stopCh       chan struct{}
	mu           sync.Mutex
	running      bool
}

func New(cfg *config.Config, ghClient *gh.Client, s *state.Store, handler NotifHandler, errHandler ErrorHandler) *Poller {
	return &Poller{
		cfg:        cfg,
		gh:         ghClient,
		state:      s,
		handler:    handler,
		errHandler: errHandler,
		stopCh:     make(chan struct{}),
	}
}

func (p *Poller) Start() {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return
	}
	p.running = true
	p.stopCh = make(chan struct{})
	p.mu.Unlock()

	go func() {
		time.Sleep(initialPollDelay)
		p.poll()

		ticker := time.NewTicker(time.Duration(p.cfg.PollIntervalMin) * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.poll()
			case <-p.stopCh:
				return
			}
		}
	}()
}

func (p *Poller) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.running {
		p.running = false
		close(p.stopCh)
	}
}

func (p *Poller) poll() {
	now := time.Now()
	if !p.shouldPoll(now) {
		return
	}

	lastPoll := p.state.GetLastPollTime()
	notifs, err := p.gh.GetNotifications(lastPoll)
	if err != nil {
		log.Printf("poll error: %v", err)
		if p.errHandler != nil {
			p.errHandler(err.Error())
		}
		return
	}

	var newNotifs []gh.Notification
	ids := make([]string, 0, len(notifs))
	for _, n := range notifs {
		ids = append(ids, n.ID)
		if p.state.IsNew(n.ID) {
			newNotifs = append(newNotifs, n)
		}
	}
	p.state.MarkSeenBatch(ids)
	p.state.SetLastPollTime(now)
	if err := p.state.Save(); err != nil {
		log.Printf("state save error: %v", err)
	}

	if len(newNotifs) > 0 {
		if p.cfg.SoundEnabled {
			if err := sound.Play(p.cfg.SoundPath); err != nil {
				log.Printf("sound error: %v", err)
			}
		}
		if p.handler != nil {
			p.handler(newNotifs)
		}
	}
}

func (p *Poller) shouldPoll(now time.Time) bool {
	day := now.Weekday().String()
	hour := now.Hour()

	if len(p.cfg.ScheduleDays) == 0 {
		return true
	}

	dayOk := false
	for _, d := range p.cfg.ScheduleDays {
		if strings.EqualFold(d, day) {
			dayOk = true
			break
		}
	}
	if !dayOk {
		return false
	}

	return hour >= p.cfg.ScheduleStartHour && hour < p.cfg.ScheduleEndHour
}
