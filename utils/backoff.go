package utils

import (
	"math"
	"math/rand"
	"sync/atomic"
	"time"
)

const (
	DefaultMin    time.Duration = 100 * time.Millisecond
	DefaultMax    time.Duration = 10 * time.Second
	DefaultFactor float64       = 2

	maxInt64 = float64(math.MaxInt64 - 512)
)

type Backoff struct {
	attempt uint64
	factor  float64
	jitter  bool
	min     time.Duration
	max     time.Duration
}

func NewBackoff(min time.Duration, max time.Duration, factor float64, jitter bool) *Backoff {
	return &Backoff{
		attempt: 0,
		factor:  factor,
		jitter:  jitter,
		min:     min,
		max:     max,
	}
}

// NewDefaultBackoff return *Backoff with default values
// min defaults to 100 milliseconds.
// max defaults to 10 seconds.
// factor defaults to 2.
// jitter defaults to false.
func NewDefaultBackoff() *Backoff {
	return NewBackoff(DefaultMin, DefaultMax, DefaultFactor, false)
}

func (b *Backoff) Duration() time.Duration {
	d := b.ForAttempt(float64(atomic.AddUint64(&b.attempt, 1) - 1))
	return d
}

func (b *Backoff) ForAttempt(attempt float64) time.Duration {
	minDuration := b.min
	if minDuration <= 0 {
		minDuration = DefaultMin
	}

	maxDuration := b.max
	if maxDuration <= 0 {
		maxDuration = DefaultMax
	}

	if minDuration >= maxDuration {
		return maxDuration
	}

	factor := b.factor
	if factor <= 0 {
		factor = DefaultFactor
	}

	minf := float64(minDuration)
	durf := minf * math.Pow(factor, attempt)

	if b.jitter {
		durf = rand.Float64()*(durf-minf) + minf
	}
	// bug with overflow to Duration  - int64
	if durf > maxInt64 {
		return maxDuration
	}
	dur := time.Duration(durf)

	//keep within bounds
	if dur < minDuration {
		return minDuration
	}
	if dur > maxDuration {
		return maxDuration
	}
	return dur
}

func (b *Backoff) Reset() {
	atomic.StoreUint64(&b.attempt, 0)
}

func (b *Backoff) Attempt() uint64 {
	return atomic.LoadUint64(&b.attempt)
}
