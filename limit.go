package limit

import (
	"time"

	"github.com/wzshiming/buffer"
)

type Limit struct {
	buf *buffer.Buffer
	max map[string]uint64
}

func NewLimit() *Limit {
	return &Limit{
		buf: buffer.NewBuffer(),
		max: map[string]uint64{},
	}
}

func (l *Limit) Limit(count uint64, s time.Duration, k ...string) bool {
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
