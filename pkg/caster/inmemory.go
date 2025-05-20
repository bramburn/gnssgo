package caster

import (
	"context"
	"io"
	"sync"
)

// InMemorySourceService is a simple in-memory implementation of SourceService
type InMemorySourceService struct {
	Sourcetable Sourcetable
	mutex       sync.RWMutex
	mounts      map[string]*mountPoint
}

// mountPoint represents a mount point in the in-memory source service
type mountPoint struct {
	name        string
	subscribers []chan []byte
	mutex       sync.RWMutex
}

// NewInMemorySourceService creates a new in-memory source service
func NewInMemorySourceService() *InMemorySourceService {
	return &InMemorySourceService{
		mounts: make(map[string]*mountPoint),
	}
}

// GetSourcetable returns the sourcetable
func (s *InMemorySourceService) GetSourcetable() Sourcetable {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Sourcetable
}

// Publisher creates a new publisher for the given mountpoint
func (s *InMemorySourceService) Publisher(ctx context.Context, mount, username, password string) (io.WriteCloser, error) {
	s.mutex.Lock()
	mp, ok := s.mounts[mount]
	if !ok {
		mp = &mountPoint{
			name:        mount,
			subscribers: make([]chan []byte, 0),
		}
		s.mounts[mount] = mp
	}
	s.mutex.Unlock()

	return &publisher{
		ctx:    ctx,
		mount:  mp,
		svc:    s,
		closed: false,
	}, nil
}

// Subscriber creates a new subscriber for the given mountpoint
func (s *InMemorySourceService) Subscriber(ctx context.Context, mount, username, password string) (chan []byte, error) {
	s.mutex.RLock()
	mp, ok := s.mounts[mount]
	if !ok {
		s.mutex.RUnlock()
		return nil, ErrorNotFound
	}
	s.mutex.RUnlock()

	mp.mutex.Lock()
	ch := make(chan []byte, 10)
	mp.subscribers = append(mp.subscribers, ch)
	mp.mutex.Unlock()

	// Remove the subscriber when the context is done
	go func() {
		<-ctx.Done()
		mp.mutex.Lock()
		for i, sub := range mp.subscribers {
			if sub == ch {
				mp.subscribers = append(mp.subscribers[:i], mp.subscribers[i+1:]...)
				break
			}
		}
		mp.mutex.Unlock()
		close(ch)
	}()

	return ch, nil
}

// publisher implements io.WriteCloser for publishing data to subscribers
type publisher struct {
	ctx    context.Context
	mount  *mountPoint
	svc    *InMemorySourceService
	closed bool
	mutex  sync.Mutex
}

// Write writes data to all subscribers
func (p *publisher) Write(data []byte) (int, error) {
	p.mutex.Lock()
	if p.closed {
		p.mutex.Unlock()
		return 0, io.ErrClosedPipe
	}
	p.mutex.Unlock()

	// Check if the context is done
	select {
	case <-p.ctx.Done():
		return 0, p.ctx.Err()
	default:
	}

	// Copy the data to avoid race conditions
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// Send the data to all subscribers
	p.mount.mutex.RLock()
	for _, sub := range p.mount.subscribers {
		select {
		case sub <- dataCopy:
		default:
			// Skip if the channel is full
		}
	}
	p.mount.mutex.RUnlock()

	return len(data), nil
}

// Close closes the publisher
func (p *publisher) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.closed = true
	return nil
}
