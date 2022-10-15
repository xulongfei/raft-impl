package raft

import (
	"log"
	"net/rpc/jsonrpc"
	"time"
)

type Node struct {
	Id     int
	Dir    string
	Status int

	//Persistent state on all servers:
	Entries     []*LogEntry
	VotedFor    int //
	CurrentTerm int

	//Volatile state on all servers:
	CommitIndex int
	LastApplied int

	//Volatile state on leaders
	NextIndex    []int
	MatchIndex   []int
	Members      Members
	electionTime time.Time
}

func NewNode(id int, dir string, memberList string) (*Node, error) {
	members, err := parseMemberList(memberList)
	if err != nil {
		return nil, err
	}
	n := &Node{
		Id:       id,
		Dir:      dir,
		Members:  members,
		Status:   follower,
		VotedFor: -1,
	}
	n.load()
	go n.loop()
	log.Println("node started")
	return n, nil
}

func (n *Node) load() {

}

func (n *Node) loop() {
	electionTicker := time.NewTicker(electionTimeout)
	heartbeatTicker := time.NewTicker(heartbeatTimeout)
	for {
		select {
		case <-electionTicker.C:
			elapsed := time.Now().Sub(n.electionTime)
			log.Println("loop electionTimeout:", elapsed.String())
			if elapsed >= electionTimeout {
				n.beFollower()
			}
		case <-heartbeatTicker.C:
			n.beLeader()
		}
	}
}

// LastLog   index,term
func (n *Node) LastLog() (int, int) {
	if len(n.Entries) == 0 {
		return 0, 0
	}
	entry := n.Entries[len(n.Entries)-1]
	return entry.Index, entry.Term
}

func (n *Node) clear() {
	n.VotedFor = -1
	n.Status = follower
}

func (n *Node) do() {
	switch n.Status {
	case follower:
		n.beLeader()
	case candidate:
		n.beCandidate()
	case leader:
		n.beLeader()
	}
}

func (n *Node) resetElectionTimeout() {
	n.electionTime = time.Now()
}

func (n *Node) beFollower() {
	if n.Status != follower {
		return
	}
	time.AfterFunc(getRandomMilliSec(100, 300), func() {
		n.Status = candidate
		n.resetElectionTimeout()
		n.do()
		log.Println("change to be candidate")
	})
}

func (n *Node) beCandidate() {
	if n.Status != candidate {
		return
	}
	n.CurrentTerm++ //先自增
	n.Members.WhenAll(n.reqVote, func(results map[int]interface{}) {
		vote := 1 //先为自己投1票
		for _, ret := range results {
			if ret == nil {
				continue
			}
			if rsp, ok := ret.(*VoteRsp); ok && rsp != nil {
				if rsp.VoteGranted {
					vote++
					if vote >= n.Members.Quorum() { //赢得选举
						n.Status = leader
						n.do()
						log.Println("vote success,change to be leader")
						return
					}
				}
			}
		}
		n.Status = follower
		log.Println("vote failed,change to be follower")
	})
}

func (n *Node) beLeader() {
	if n.Status != leader {
		return
	}
	n.Members.WhenAll(n.reqAppendEntries, func(results map[int]interface{}) {
		//复制日志
	})
}

func (n *Node) reqVote(m *Member) interface{} {
	if n.Id == m.Id {
		return nil
	}
	conn, err := jsonrpc.Dial("tcp", m.Host)
	if err != nil {
		log.Println("dial failed:", err)
		return nil
	}
	defer conn.Close()
	rsp := &VoteRsp{}
	lastLogIndex, lastLogTerm := n.LastLog()
	err = conn.Call("raft.HandleVote", &VoteReq{
		Term:         n.CurrentTerm,
		CandidateId:  n.Id,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}, rsp)
	if err != nil {
		return nil
	}
	log.Printf("reqVote to member:%v,rsp:%v \n", m, rsp)
	return rsp
}

func (n *Node) reqAppendEntries(m *Member) interface{} {
	if n.Id == m.Id {
		return nil
	}
	conn, err := jsonrpc.Dial("tcp", m.Host)
	if err != nil {
		log.Println("dial failed:", err)
		return nil
	}
	defer conn.Close()
	rsp := &AppendEntriesRsp{}
	err = conn.Call("raft.HandleAppendEntries", &AppendEntriesReq{
		Term:     n.CurrentTerm,
		LeaderId: n.Id,
	}, rsp)
	if err != nil {
		return nil
	}
	log.Printf("reqAppendEntries to member:%v,rsp:%v \n", m, rsp)
	return rsp
}

func (n *Node) logUpToDate(req *VoteReq) bool {
	idx, term := n.LastLog()
	return idx <= req.LastLogIndex && term <= req.LastLogTerm
}

func (n *Node) HandleVote(req *VoteReq, rsp *VoteRsp) error {
	log.Printf("recv vote request:%v", req)
	rsp.Term = n.CurrentTerm
	if req.Term < n.CurrentTerm {
		rsp.VoteGranted = false
		log.Println("recv vote, reject")
		return nil
	}
	if ((n.VotedFor == -1 || n.VotedFor == req.CandidateId) && n.logUpToDate(req)) ||
		req.Term > n.CurrentTerm {
		rsp.VoteGranted = true
		n.VotedFor = req.CandidateId
		n.Status = follower
		n.CurrentTerm = req.Term
		n.resetElectionTimeout()
		log.Println("recv vote, change to be follower")
	}
	return nil
}

func (n *Node) HandleAppendEntries(req *AppendEntriesReq, rsp *AppendEntriesRsp) error {
	log.Printf("recv appendEntries request:%v \n", req)
	rsp.Term = n.CurrentTerm
	if n.CurrentTerm > req.Term {
		rsp.Success = false
		log.Println("recv appendEntries, reject")
		return nil
	}
	if req.Term >= n.CurrentTerm {
		rsp.Success = true
		n.Status = follower
		n.CurrentTerm = req.Term
		n.resetElectionTimeout()
		log.Println("recv appendEntries,change to follower")
	}
	return nil
}
