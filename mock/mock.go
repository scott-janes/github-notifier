package mock

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/scott-janes/github-notifier/gh"
)

type Handler func([]gh.Notification)

type Poller struct {
	handler Handler
	stopCh  chan struct{}
}

func New(handler Handler) *Poller {
	return &Poller{
		handler: handler,
		stopCh:  make(chan struct{}),
	}
}

func (p *Poller) Start() {
	go func() {
		time.Sleep(2 * time.Second)
		p.emit()

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.emit()
			case <-p.stopCh:
				return
			}
		}
	}()
}

func (p *Poller) Stop() {
	select {
	case <-p.stopCh:
	default:
		close(p.stopCh)
	}
}

var mockItems = []struct {
	repo, title, reason string
}{
	{"owner/frontend-app", "Fix login button alignment on mobile", "review_requested"},
	{"owner/frontend-app", "Add dark mode support", "mention"},
	{"owner/api-service", "Optimize database query for user search", "author"},
	{"owner/docs-repo", "Update API docs for v2 endpoints", "comment"},
	{"owner/infra-tools", "Add kubernetes deployment scripts", "state_change"},
	{"owner/design-system", "Create new Button component variants", "review_requested"},
	{"owner/cli-tool", "Add --verbose flag to build command", "mention"},
	{"owner/mobile-app", "Fix push notification registration", "author"},
	{"owner/frontend-app", "Refactor auth middleware to use JWT", "review_requested"},
	{"owner/api-service", "Add rate limiting per user tier", "comment"},
	{"owner/docs-repo", "Add migration guide for v1 to v2", "author"},
	{"owner/infra-tools", "Upgrade CI pipeline to use ARM runners", "mention"},
	{"owner/design-system", "Add Toast notification component", "review_requested"},
	{"owner/cli-tool", "Support JSON output in list command", "state_change"},
}

func (p *Poller) emit() {
	n := rand.Intn(3) + 1
	now := time.Now()
	notifs := make([]gh.Notification, n)

	perm := rand.Perm(len(mockItems))
	for i := 0; i < n; i++ {
		item := mockItems[perm[i]]
		notifs[i] = gh.Notification{
			ID:        fmt.Sprintf("mock-%d-%d", now.UnixNano(), i),
			Reason:    item.reason,
			Title:     item.title,
			Repo:      item.repo,
			URL:       fmt.Sprintf("https://github.com/%s/pull/%d", item.repo, rand.Intn(500)+1),
			UpdatedAt: now.Add(-time.Duration(rand.Intn(60)) * time.Minute),
		}
	}

	if p.handler != nil {
		p.handler(notifs)
	}
}
