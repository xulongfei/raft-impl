package raft

import "time"

const (
	follower = iota
	candidate
	leader
)

const (
	electionTimeout  = time.Second * 5
	heartbeatTimeout = time.Millisecond * 3000
)
