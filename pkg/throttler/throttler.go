package throttler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Throttler struct {
	m       sync.RWMutex
	ctx     context.Context
	credit  float64
	lastHit time.Time
}
type throttlerState struct {
	Credit  float64   `json:"credit"`
	LastHit time.Time `json:"lastHit"`
}

func NewThrottler(ctx context.Context) *Throttler {
	return &Throttler{
		m:      sync.RWMutex{},
		ctx:    ctx,
		credit: 100,
	}
}
func NewThrottlerFrom(ctx context.Context, serializedState string) (*Throttler, error) {
	var state throttlerState
	if json.Valid([]byte(serializedState)) {
		var state throttlerState
		if err := json.Unmarshal([]byte(serializedState), &state); err != nil {
			return nil, err
		}
	}
	state.Init()
	return &Throttler{
		m:      sync.RWMutex{},
		ctx:    ctx,
		credit: state.Credit,
	}, nil
}

func (t *Throttler) Initialize() {
	go func() {
		for {
			func() {
				inc := rand.Float64() * 15
				t.m.Lock()
				if inc+t.credit > 100 {
					t.credit = 100
				} else {
					t.credit += inc
				}
				t.m.Unlock()
			}()

			select {
			case <-t.ctx.Done():
				return
			default:
				time.Sleep(time.Millisecond * 500)
				continue
			}
		}
	}()
}

func (t *Throttler) Work(title string) (time.Duration, error) {
	// Let's do some sync work
	// Makes some randomly workCost, max workCost is at 100
	workCost := 100 * rand.Float64()

	processStart := time.Now()

	// Checks if we have credit to do the work if not we wait
	if err := repeatTillCancel(t.ctx, func() (bool, error) {
		t.m.RLock()
		if t.credit >= workCost {
			log.Printf("ready credit %s -> cost %f -- credit %f", title, workCost, t.credit)
			t.m.RUnlock()
			return true, nil
		}
		log.Printf("waiting for credits %s -> cost %f -- credit %f", title, workCost, t.credit)
		t.m.RUnlock()

		// Waits for some time
		time.Sleep(time.Second * 2)

		// Continue waiting, returns false to repeat the work
		return false, nil
	}); err != nil {
		return time.Since(processStart), err
	}

	// Locks up the mutex, we're going to do the work
	t.m.Lock()
	defer t.m.Unlock()

	executionStart := time.Now()
	if err := repeatTillCancel(t.ctx, func() (bool, error) {
		if time.Since(executionStart) >= time.Second {
			return true, nil
		}
		time.Sleep(time.Millisecond * 1)
		return false, nil
	}); err != nil {
		return time.Since(processStart), err
	}

	// Updates current credits
	t.credit -= workCost

	// Exists
	return time.Since(processStart), nil
}

func (t *Throttler) Serialize() string {
	b, _ := json.Marshal(throttlerState{Credit: t.credit, LastHit: t.lastHit})
	return string(b)
}

func (s *throttlerState) Init() {

}

func repeatTillCancel(ctx context.Context, fn func() (done bool, err error)) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("cancelled")
		default:
			if done, err := fn(); done {
				return err
			}
			time.Sleep(time.Nanosecond * 100)
		}
	}
}
