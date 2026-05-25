package worker

import (
	"time"
)


type backoff struct {
	initial time.Duration
	max 	time.Duration
	current time.Duration
}

func NewBackoff(initial, max time.Duration) *backoff {
	return &backoff{
		initial: initial,
		max: max,
		current: initial,
	}
}

func (b *backoff) Next() time.Duration {
	d := b.current

	b.current *= 2
	
	if b.current > b.max {
		b.current = b.max
	}

	return d
}

func (b *backoff) Reset(){
	b.current = b.initial
}