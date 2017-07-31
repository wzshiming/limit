package limit

import (
	"time"

	"sync"

	"github.com/wzshiming/buffer"
	"github.com/wzshiming/task"
)

type Limit struct {
	buf   *buffer.Buffer
	max   map[string]uint64
	reset *task.Task
	mux   sync.RWMutex
}

func NewLimit(f func() time.Time) *Limit {
	l := &Limit{
		buf: buffer.NewBuffer(),
		max: map[string]uint64{},
	}

	if f != nil {
		l.reset = task.NewTask(1)
		l.reset.AddPeriodic(f, func() {
			l.mux.Lock()
			defer l.mux.Unlock()
			l.buf = buffer.NewBuffer()
			l.max = map[string]uint64{}
		})
	}
	return l
}

func (l *Limit) Limit(count uint64, s time.Duration, k ...string) bool {
	l.mux.RLock()
	defer l.mux.RUnlock()
	pass := false
	ks := []func() (interface{}, time.Time, error){}
	run := func() (interface{}, time.Time, error) {
		ff := ks[len(ks)-1]
		ks = ks[:len(ks)-1]
		return ff()
	}
	for _, v := range k {
		func(v string) {
			ff := func() (interface{}, time.Time, error) {
				return l.buf.Buf(v, func() (interface{}, time.Time, error) {
					if len(ks) != 0 {
						return run()
					} else {
						pass = true
						return nil, time.Now().Add(s), nil
					}
				})
			}
			ks = append(ks, ff)
		}(v)
	}
	run()
	if pass {
		for _, v := range k {
			if l.max[v] >= uint64(count) {
				pass = false
			} else {
				l.max[v]++
			}
		}
	}
	return pass
}
