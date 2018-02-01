package raft

import (
	
)

/*
 * This is interface to be used by upper layer/testing program in co-ordination
 * with wrappers of Action and State machine. 
 * */

/* Possible bug: using pointer inside struct might lead to dangling pointer problem.*/

func MakeAppendEvent(data []byte) *Event{
	
	//Basically create Event object with appropriate pointer to Data structs set.
	//Other pointers will be NIL
	e := Event{EventType : "append", From : 0, Term : 0, Append: &AppendData{Data:data}}
 	return &e
}

func MakeHeartbeatEvent(from int,term int,leaderId int,leaderCommit int) *Event{
	
	e := Event{EventType: "appendReq",From: from, Term: term, AppendReq: &AppendEntriesReqData{LeaderId:leaderId,PrevLogIndex: -1, PrevLogTerm: -1, Entries: nil,LeaderCommit: leaderCommit}}
	
	return &e
}

func MakeAppendEntriesReqEvent(from int,term int,leaderId int, prevLogIndex int,prevLogTerm int,entries []*LogEntry,leaderCommit int) *Event{
	
	e := Event{EventType: "appendReq",From: from, Term: term, AppendReq: &AppendEntriesReqData{LeaderId:leaderId,PrevLogIndex: prevLogIndex, PrevLogTerm: prevLogTerm, Entries: entries,LeaderCommit: leaderCommit}}
	
	return &e
}

func MakeAppendEntriesRespEvent(from int, term int, success bool,index int) *Event{
	
	e := Event{EventType: "appendResp", From: from, Term: term, AppendResp: &AppendEntriesRespData{Success:success,Index:index}}
	return &e
	
}

func MakeTimeoutEvent() *Event{
	
	e := Event{EventType: "timeout", From : 0, Term : 0, Timeout: &TimeoutData{}}
	return &e
	
}

func MakeStopEvent() *Event{
	
	e := Event{EventType: "stop", From : 0, Term : 0}
	return &e
	
}

func MakeVoteReqEvent(from int, term int, candidateId int, lastLogIndex int, lastLogTerm int) *Event{
	
	e := Event{EventType: "voteReq", From: from, Term: term, VoteReq: &VoteReqData{CandidateId:candidateId,  LastLogIndex:lastLogIndex, LastLogTerm:lastLogTerm}}
	return &e
	
}

func MakeVoteRespEvent(from int,term int, voteGranted bool) *Event{
	e := Event{EventType: "voteResp", From: from, Term: term, VoteResp: &VoteRespData{VoteGranted: voteGranted}}
	return &e
}
