package raft

type LogEntry struct {
	Term  int
	Index int
	Cmd   []byte //SET DEL INCR
	Key   []byte
	Val   []byte
}
