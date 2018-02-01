package raft

/*import (
	"strings"
)*/

/*
 * This file contains implementation of Events.
 * Only one event type is defined, so that a single communication
 * channel can be used between client and raft state machine.
 * Further to impose modularity and space efficiency, smaller structures
 * corresponding to event types are defined and main event object
 * contains pointers to these objects.
 * 
 * */

type AppendData struct {
	
		Data []byte
}

type AppendEntriesReqData struct {
	
	LeaderId int
	PrevLogIndex int
	PrevLogTerm int
	//Entries []byte
	Entries []*LogEntry
	LeaderCommit int

}

type AppendEntriesRespData struct {
	
	Success bool
	Index int

}

type TimeoutData struct {
	
	//Time int
	
}
type VoteReqData struct {
	
	CandidateId int
	LastLogIndex int
	LastLogTerm int
	
}

type VoteRespData struct {
	
	VoteGranted bool 	
		
}

type Event struct {
	
	EventType string
	From int
	
	Term int
	
	Append *AppendData
	AppendReq *AppendEntriesReqData
	AppendResp *AppendEntriesRespData
	Timeout *TimeoutData
	VoteReq *VoteReqData
	VoteResp *VoteRespData
}
