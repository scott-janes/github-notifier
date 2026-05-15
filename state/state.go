package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/scott-janes/github-notifier/config"
)

type Store struct {
	mu           sync.RWMutex
	seenIDs      map[string]bool
	confirmedIDs map[string]bool
	lastPollTime time.Time
	filePath     string
}

type persistedState struct {
	SeenIDs      []string  `json:"seen_ids"`
	ConfirmedIDs []string  `json:"confirmed_ids"`
	LastPollTime time.Time `json:"last_poll_time"`
}

func New() *Store {
	s := &Store{
		seenIDs:      make(map[string]bool),
		confirmedIDs: make(map[string]bool),
		filePath:     filepath.Join(config.DefaultConfigDir(), "state.json"),
	}
	s.load()
	return s
}

func (s *Store) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}
	var ps persistedState
	if err := json.Unmarshal(data, &ps); err != nil {
		return
	}
	for _, id := range ps.SeenIDs {
		s.seenIDs[id] = true
	}
	for _, id := range ps.ConfirmedIDs {
		s.confirmedIDs[id] = true
	}
	s.lastPollTime = ps.LastPollTime
}

func (s *Store) Save() error {
	s.mu.Lock()
	ps := persistedState{
		LastPollTime: s.lastPollTime,
	}
	for id := range s.seenIDs {
		ps.SeenIDs = append(ps.SeenIDs, id)
	}
	for id := range s.confirmedIDs {
		ps.ConfirmedIDs = append(ps.ConfirmedIDs, id)
	}
	s.mu.Unlock()

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Store) IsNew(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return !s.seenIDs[id]
}

func (s *Store) MarkSeen(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seenIDs[id] = true
}

func (s *Store) Confirm(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.confirmedIDs[id] = true
}

func (s *Store) ConfirmAll(ids []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		s.confirmedIDs[id] = true
	}
}

func (s *Store) GetLastPollTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastPollTime
}

func (s *Store) SetLastPollTime(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastPollTime = t
}

func (s *Store) Confirmed(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.confirmedIDs[id]
}

func (s *Store) MarkSeenBatch(ids []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		s.seenIDs[id] = true
	}
}
