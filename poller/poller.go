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
	cfg        *config.Config
	gh         *gh.Client
	state      *state.Store
	handler    NotifHandler
	errHandler ErrorHandler
	afterPoll  func()
	stopCh     chan struct{}
	mu         sync.Mutex
	running    bool
}

func New(cfg *config.Config, ghClient *gh.Client, s *state.Store, handler NotifHandler, errHandler ErrorHandler, afterPoll func()) *Poller {
	return &Poller{
		cfg:        cfg,
		gh:         ghClient,
		state:      s,
		handler:    handler,
		errHandler: errHandler,
		afterPoll:  afterPoll,
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
		if p.afterPoll != nil {
			p.afterPoll()
		}
		return
	}

	lastPoll := p.state.GetLastPollTime()
	notifs, err := p.gh.GetNotifications(lastPoll)
	if err != nil {
		log.Printf("poll error: %v", err)
		if p.errHandler != nil {
			p.errHandler(err.Error())
		}
		if p.afterPoll != nil {
			p.afterPoll()
		}
		return
	}

	var newNotifs []gh.Notification
	keys := make([]string, 0, len(notifs))
	for _, n := range notifs {
		key := n.ID + "@" + n.UpdatedAt.UTC().Format(time.RFC3339Nano)
		keys = append(keys, key)
		if p.state.IsNew(key) {
			newNotifs = append(newNotifs, n)
		}
	}
	p.state.MarkSeenBatch(keys)
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

	if p.afterPoll != nil {
		p.afterPoll()
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
