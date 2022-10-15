package raft

import (
	"errors"
	"strconv"
	"strings"
	"sync"
)

var ErrInvalidMember = errors.New("invalid member")

type Member struct {
	Id   int
	Host string //ip:port
}

func newMember(conf string) (*Member, error) {
	segments := strings.Split(conf, ",")
	if len(segments) != 2 {
		return nil, ErrInvalidMember
	}
	id, err := strconv.Atoi(segments[0])
	if err != nil {
		return nil, err
	}
	return &Member{
		Id:   id,
		Host: segments[1],
	}, nil
}

type Members map[int]*Member

func (m Members) GetMember(id int) *Member {
	return m[id]
}

func (m Members) Quorum() int {
	return len(m)/2 + 1
}

// 0,127.0.0.1:9001;1,127.0.0.1:9002;2,127.0.0.1:9003
func parseMemberList(memberList string) (Members, error) {
	ret := Members{}
	members := strings.Split(memberList, ";")
	for _, m := range members {
		member, err := newMember(m)
		if err != nil {
			return nil, err
		}
		ret[member.Id] = member
	}
	return ret, nil
}

func (m Members) WhenAll(f func(member *Member) interface{}, done func(map[int]interface{})) {
	var wg sync.WaitGroup
	var results = map[int]interface{}{}
	var mut sync.Mutex
	for _, member := range m {
		wg.Add(1)
		go func(member *Member) {
			defer wg.Done()
			rsp := f(member)
			mut.Lock()
			results[member.Id] = rsp
			mut.Unlock()
		}(member)
	}
	wg.Wait()
	done(results)
}
