package raft

import (
	"fmt"
	"testing"
	"os"
)


func TestCandidateAppendReq1(t *testing.T){
	//Older term including prev term
	
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * AppendReq: from,term,leaderId,prevLogIndex,prevLogTerm,entries,
	 * leaderCommit*/
	 
	sm := InitGenericSM()
	sm.CurrentState = "candidate"
	sm.VotedFor = 1 //voted to self
	sm.CurrentTerm = 6
	
	var entriesToSend []*LogEntry
	
	ev := MakeAppendEntriesReqEvent(4,5,4,3,4,append(entriesToSend,&LogEntry{5,[]byte("abc")}),4)
	a := sm.RouteEvent(ev)
	
	ev1 := a[0].Send.EventToSend
	
	if(a[0].ActionType != "send" || a[0].Send.PeerId != ev.From) {
		t.Errorf("Error1")
	}
	
	if(ev1.AppendResp.Success != false || ev1.Term != sm.CurrentTerm){
		t.Errorf("Error2")
	}

	os.RemoveAll("SMLogs")
}

func TestCandidateTimeout1(t *testing.T){
	//No majority
	
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * AppendReq: from,term,leaderId,prevLogIndex,prevLogTerm,entries,
	 * leaderCommit*/
	 
	sm := InitGenericSM()
	sm.CurrentState = "candidate"
	sm.VotedFor = 1 //voted to self
	sm.CurrentTerm = 6
	sm.VoteCounter=2
	
	ev := MakeTimeoutEvent()

	a := sm.RouteEvent(ev)
	
	if(a[0].ActionType != "alarm" || sm.CurrentState != "follower") {
		t.Errorf("Error1")
	}
	os.RemoveAll("SMLogs")
}

func TestCandidateTimeout2(t *testing.T){
	//Majority was obtained
	
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * AppendReq: from,term,leaderId,prevLogIndex,prevLogTerm,entries,
	 * leaderCommit*/
	 
	sm := InitGenericSM()
	sm.CurrentState = "candidate"
	sm.VotedFor = 1 //voted to self
	sm.CurrentTerm = 6
	sm.VoteCounter=3
	
	ev := MakeTimeoutEvent()

	a := sm.RouteEvent(ev)
	
	if(a[0].ActionType != "alarm" || sm.CurrentState != "leader") {
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")
}

func TestCandidateTimeout3(t *testing.T){
	//Too many negative votes
	
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * AppendReq: from,term,leaderId,prevLogIndex,prevLogTerm,entries,
	 * leaderCommit*/
	 
	sm := InitGenericSM()
	sm.CurrentState = "candidate"
	sm.VotedFor = 1 //voted to self
	sm.CurrentTerm = 6
	sm.VoteCounter=1
	sm.NegativeVoteCounter = 4
	
	ev := MakeTimeoutEvent()

	a := sm.RouteEvent(ev)
	
	if(a[0].ActionType != "alarm" || sm.CurrentState != "follower") {
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")
}

func TestCandidateVoteReq1(t *testing.T){
	//VoteReq for higher term
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteReq: from int, term int, candidateId int, lastLogIndex int, lastLogTerm int*/
	 
	sm := InitGenericSM()
	sm.CurrentState = "candidate"
	sm.VotedFor = 1 //voted to self
	sm.CurrentTerm = 6
	sm.VoteCounter=3
	
	ev := MakeVoteReqEvent(4,7,3,4,4)

	sm.RouteEvent(ev)
	//voteReq for follower already tesed
	if(sm.CurrentState != "follower") {
		fmt.Println(sm.CurrentState)
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")
}

func TestCandidateVoteReq2(t *testing.T){
	//VoteReq for lower term
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteReq: from int, term int, candidateId int, lastLogIndex int, lastLogTerm int*/
	 
	sm := InitGenericSM()
	sm.CurrentState = "candidate"
	sm.VotedFor = 1 //voted to self
	sm.CurrentTerm = 6
	sm.VoteCounter=3
	
	ev := MakeVoteReqEvent(4,5,4,4,4)

	a := sm.RouteEvent(ev)
	
	ev1 := a[0].Send.EventToSend
	
	if(sm.CurrentState != "candidate" || ev1.VoteResp.VoteGranted != false || a[0].Send.PeerId != ev.From || ev1.Term != sm.CurrentTerm) {
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")
}

func TestCandidateVoteResp1(t *testing.T){
	//VoteResp voteGranted for this term, no majority
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteResp: from int, term int, voteGranted*/
	 
	sm := InitGenericSM()
	sm.CurrentState = "candidate"
	sm.VotedFor = 1 //voted to self
	sm.CurrentTerm = 6
	sm.VoteCounter=1
	
	ev := MakeVoteRespEvent(3,6,true)
	
	sm.RouteEvent(ev)
	
	if(sm.VoteCounter != 2 && sm.CurrentState != "candidate") {
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")	
}

func TestCandidateVoteResp2(t *testing.T){
	//VoteResp voteGranted for this term, majority
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteResp: from int, term int, voteGranted*/
	 
	sm := InitGenericSM()
	sm.CurrentState = "candidate"
	sm.VotedFor = 1 //voted to self
	sm.CurrentTerm = 6
	sm.VoteCounter=2
	
	ev := MakeVoteRespEvent(3,6,true)
	
	a := sm.RouteEvent(ev)
	
	if(sm.VoteCounter != 3 && sm.CurrentState != "leader") {
		t.Errorf("Error")
	}
	
	if(a[0].ActionType != "alarm"){
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")
}

func TestCandidateVoteResp3(t *testing.T){
	//VoteResp voteGranted false, higher term
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteResp: from int, term int, voteGranted*/
	 
	sm := InitGenericSM()
	sm.CurrentState = "candidate"
	sm.VotedFor = 1 //voted to self
	sm.CurrentTerm = 6
	sm.VoteCounter=2
	
	ev := MakeVoteRespEvent(3,7,false)
	
	a := sm.RouteEvent(ev)
	
	if(sm.CurrentState != "follower") {
		t.Errorf("Error")
	}
	
	if(a[0].ActionType != "alarm"){
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")
}

func TestCandidateVoteResp4(t *testing.T){
	//VoteResp voteGranted false, same term, NegetiveVoteCounter exceeds majority
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteResp: from int, term int, voteGranted*/
	 
	sm := InitGenericSM()
	sm.CurrentState = "candidate"
	sm.VotedFor = 1 //voted to self
	sm.CurrentTerm = 6
	sm.VoteCounter=1
	sm.NegativeVoteCounter=3
	
	ev := MakeVoteRespEvent(3,6,false)
	
	a := sm.RouteEvent(ev)
	
	if(sm.CurrentState != "follower") {
		t.Errorf("Error")
	}
	
	if(a[0].ActionType != "alarm" && a[0].Alarm.Timeout != ELECTION_TIMEOUT ){
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")
}

func TestCandidateVoteResp5(t *testing.T){
	//VoteResp voteGranted false, same term,NegativeVoteCounter doesn't exceed majority
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteResp: from int, term int, voteGranted*/
	 
	sm := InitGenericSM()
	sm.CurrentState = "candidate"
	sm.VotedFor = 1 //voted to self
	sm.CurrentTerm = 6
	sm.VoteCounter=1
	sm.NegativeVoteCounter=2
	
	ev := MakeVoteRespEvent(3,6,false)
	
	sm.RouteEvent(ev)
	
	if(sm.CurrentState != "candidate") {
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")	
}

