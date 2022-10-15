package main

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"raft/raft"
)

var dataDir = "."
var nodeId = 0
var memberList = ""

func main() {
	flag.StringVar(&dataDir, "dataDir", ".", "dataDir")
	flag.IntVar(&nodeId, "nodeId", 0, "nodeId")
	flag.StringVar(&memberList, "memberList", "", "memberList")
	flag.Parse()
	node, err := raft.NewNode(nodeId, dataDir, memberList)
	if err != nil {
		log.Fatal(err)
	}
	if err := rpc.RegisterName("raft", node); err != nil {
		log.Fatal(err)
	}
	lis, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		log.Fatal(err)
		return
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("err:%s \n", err.Error())
			continue
		}
		go func(conn net.Conn) {
			log.Printf("new client connected:%s \n", conn.RemoteAddr())
			jsonrpc.ServeConn(conn)
		}(conn)
	}
}
