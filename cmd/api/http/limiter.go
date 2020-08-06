package http

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type RateLimiter interface {
	Acquire() (*Token, error)
	Release(*Token)
	LimitExceededPerc() float32
	ReleaseAll()
}

type Token struct {
	ID        string
	CreatedAt time.Time
}

func NewMaxConcurrencyRateLimiter(limit int) (RateLimiter, error) {
	if limit <= 0 {
		return nil, errors.New("invalid rate limit")
	}

	m := newManager(limit)

	// Await.
	func() {
		go func() {
			for {
				select {
				case <-m.inChan:
					m.tryGenerateToken()
				case t := <-m.releaseChan:
					m.releaseToken(t)
				}
			}
		}()
	}()
	return m, nil
}

func newManager(limit int) *Manager {
	return &Manager{
		errorChan:    make(chan error),
		outChan:      make(chan *Token),
		inChan:       make(chan struct{}),
		activeTokens: make(map[string]*Token),
		releaseChan:  make(chan *Token),
		needToken:    0,
		limit:        limit,
		makeToken: func() *Token {
			return &Token{
				ID:        uuid.New().String(),
				CreatedAt: time.Now().UTC(),
			}
		},
	}
}

type Manager struct {
	errorChan    chan error
	releaseChan  chan *Token
	outChan      chan *Token
	inChan       chan struct{}
	needToken    int64
	activeTokens map[string]*Token
	tokensMutex  sync.RWMutex
	limit        int
	makeToken    func() *Token
}

func (m *Manager) Acquire() (*Token, error) {
	go func() {
		m.inChan <- struct{}{}
	}()

	select {
	case t := <-m.outChan:
		return t, nil
	case err := <-m.errorChan:
		return nil, err
	}
}

func (m *Manager) Release(t *Token) {
	go func() {
		m.releaseChan <- t
	}()
}

func (m *Manager) ReleaseAll() {
	m.tokensMutex.Lock()
	for _, t := range m.activeTokens {
		m.Release(t)
	}
	m.activeTokens = make(map[string]*Token)
	m.tokensMutex.Unlock()
}

func (m *Manager) LimitExceededPerc() float32 {
	return m.limitExceededPerc()
}

func (m *Manager) incNeedToken() {
	atomic.AddInt64(&m.needToken, 1)
}

func (m *Manager) decNeedToken() {
	atomic.AddInt64(&m.needToken, -1)
}

func (m *Manager) awaitingToken() bool {
	return atomic.LoadInt64(&m.needToken) > 0
}

func (m *Manager) limitExceededPerc() float32 {
	m.tokensMutex.RLock()
	defer m.tokensMutex.RUnlock()
	return float32(atomic.LoadInt64(&m.needToken)) / float32(m.limit)
}

func (m *Manager) isLimitExceeded() bool {
	m.tokensMutex.RLock()
	defer m.tokensMutex.RUnlock()

	if len(m.activeTokens) >= m.limit {
		return true
	}
	return false
}

func (m *Manager) releaseToken(t *Token) {
	if t == nil {
		log.Println("cannot release a nil token")
		return
	}

	m.tokensMutex.Lock()
	if _, ok := m.activeTokens[t.ID]; !ok {
		log.Printf("unable to release token %s - not in use", t)
		return
	}

	delete(m.activeTokens, t.ID)
	m.tokensMutex.Unlock()

	if m.awaitingToken() {
		m.decNeedToken()
		go m.tryGenerateToken()
	}
}

func (m *Manager) tryGenerateToken() {
	if m.isLimitExceeded() {
		m.incNeedToken()
		return
	}

	t := m.makeToken()

	m.tokensMutex.Lock()
	m.activeTokens[t.ID] = t
	m.tokensMutex.Unlock()

	go func() {
		m.outChan <- t
	}()
}
