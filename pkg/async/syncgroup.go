package async

import (
	"log"
	"sync"
	"time"
)

type SyncGroup struct {
	wg sync.WaitGroup
}

func NewSyncGroup() *SyncGroup {
	return &SyncGroup{wg: sync.WaitGroup{}}
}

func (s *SyncGroup) Add(name string) (done func()) {
	s.wg.Add(1)

	return func() {
		s.wg.Done()
	}
}

func (s *SyncGroup) AddChannel(name string, ch chan struct{}) {
	s.wg.Add(1)

	go func() {
		for {
			select {
			case <-ch:
				log.Printf("channel '%s' completed", name)
				s.wg.Done()
				return
			default:
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()
}

func (s *SyncGroup) Wait() {
	s.WaitTimeout(time.Second * 30)
}

func (s *SyncGroup) WaitTimeout(timeout time.Duration) (timedout bool) {
	toc := make(chan struct{})
	go func() {
		defer close(toc)
		s.wg.Wait()
	}()
	select {
	case <-toc:
		return false
	case <-time.After(timeout):
		return true
	}
}
