package raft

type Fsm struct {
	Data map[string]interface{}
}

func NewFsm() *Fsm {
	return &Fsm{
		Data: map[string]interface{}{},
	}
}

func (f *Fsm) Apply(key string, val interface{}) {
	f.Data[key] = val
}

func (f *Fsm) MakeSnapshot() {}
