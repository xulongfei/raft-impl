package raft

type Fsm struct {
	Data map[string]interface{}
}

func (f *Fsm) Apply(key string, val interface{}) {
	f.Data[key] = val
}
