package raft

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getRandomMilliSec(min, max int64) time.Duration {
	l := min
	r := max - min
	if r <= 0 {
		r = 1
	}
	v := l + rand.Int63n(r)
	return time.Duration(v) * time.Millisecond
}
