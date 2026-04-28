package mailer

import "sync"

type Store struct {
	mu     sync.RWMutex
	active Mailer
}

func NewStore(initial Mailer) *Store {
	return &Store{active: initial}
}

func (s *Store) Get() Mailer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

func (s *Store) Set(m Mailer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = m
}
