package raft

import(
	"testing"
	"fmt"
	"os"
//	"github.com/cs733-iitb/log"
)

func TestLeaderAppend1(t *testing.T) {
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,3)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)*/
	 
	sm := InitGenericSM()
	sm.InitLeader()
	ev := MakeAppendEvent([]byte("xyz"))
	a := sm.RouteEvent(ev)
	
	persistentLastLogTerm :=-1
	persistentLastLogIndex := sm.ThisLog.GetLastIndex()
	if(persistentLastLogIndex != -1) {
		temp,_ := sm.ThisLog.Get(persistentLastLogIndex)
		persistentLastLogTerm = temp.(LogEntry).Term
	} else{
		persistentLastLogTerm = -1
	}
	
	//fmt.Println("",persistentLastLogIndex,"+",persistentLastLogTerm)
	for i:=0;i<len(sm.KnownPeers);i++ {

		if(a[i].ActionType == "send"){
			if(int64(sm.NextIndex[a[i].Send.PeerId]) != sm.ThisLog.GetLastIndex() + 1){
				t.Errorf("Error3")
			}
			ev1 := a[i].Send.EventToSend
		
			if(ev1.From != sm.MyId || ev1.Term != sm.CurrentTerm || int64(ev1.AppendReq.PrevLogIndex) != sm.ThisLog.GetLastIndex() || ev1.AppendReq.PrevLogTerm != persistentLastLogTerm) {
			t.Errorf("Error4")
		}
		}
	}
	os.RemoveAll("SMLogs")	
	
}

func TestLeaderAppendResp1(t *testing.T) {
	//Same term, No success

	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,3)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * Append response: from,term,success*/
	 
	sm := InitGenericSM()
	sm.InitLeader()
	sm.NextIndex[3]=2
	ev := MakeAppendEntriesRespEvent(3,5,false,2)
	a := sm.RouteEvent(ev)
	
	if(sm.NextIndex[ev.From] != 1){
		t.Errorf("Error1")
	}
	var a1 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "send"){
			a1=a[i]
		}
	}
	ev1 := a1.Send.EventToSend
	
	if(a1.Send.PeerId != ev.From) {
		t.Errorf("Error2")
	}
	
	if(ev1.AppendReq.PrevLogIndex != 0 || ev1.AppendReq.PrevLogTerm != 1){
		fmt.Println(ev1.AppendReq.PrevLogIndex)
		t.Errorf("Error3")
	}
		
	os.RemoveAll("SMLogs")	

}

func TestLeaderAppendResp2(t *testing.T) {
	//Same term, success, No commit

	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,3)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * Append response: from,term,success*/
	 
	sm := InitGenericSM()
	sm.InitLeader()
	sm.NextIndex[3]=2
	sm.EntrySuccessCounter[2] = 1
	ev := MakeAppendEntriesRespEvent(3,5,true,2)
	a := sm.RouteEvent(ev)
	
	if(sm.NextIndex[ev.From] != 3){
		fmt.Println(sm.NextIndex[ev.From])
		t.Errorf("Error1")
	}

	var a1 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "send"){
			a1=a[i]
		}
	}
	ev1 := a1.Send.EventToSend
	
	//ev1 := a[0].Send.EventToSend
	
	if(a1.Send.PeerId != ev.From) {
		t.Errorf("Error2")
	}
	
	if(ev1.AppendReq.PrevLogIndex != 2 || ev1.AppendReq.PrevLogTerm != 2){
		fmt.Println(ev1.AppendReq.PrevLogIndex)
		t.Errorf("Error3")
	}
	
	os.RemoveAll("SMLogs")	

}

func TestLeaderAppendResp3(t *testing.T) {
	//Commiting entry

	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,3)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * Append response: from,term,success*/
	 
	sm := InitGenericSM()
	sm.InitLeader()

	//sm.NextIndex[3]=2Not 
	sm.EntrySuccessCounter[int(sm.ThisLog.GetLastIndex())]=3
	sm.NextIndex[3]=int(sm.ThisLog.GetLastIndex())
	ev := MakeAppendEntriesRespEvent(3,4,true,int(sm.ThisLog.GetLastIndex()))
	sm.CurrentTerm = 4
	a := sm.RouteEvent(ev)
	
	if(int64(sm.NextIndex[ev.From]) != sm.ThisLog.GetLastIndex() + 1){
		t.Errorf("Error1")
	}
	
	var a1 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "commit"){
			a1=a[i]
		}
	}

	
	if(a1== nil) {
		t.Errorf("Error2")
	}
	
	if(int64(a1.Commit.Index) != sm.ThisLog.GetLastIndex()){
		t.Errorf("Error3")
	}
	
	os.RemoveAll("SMLogs")	

}

func TestLeaderAppendResp4(t *testing.T) {
	//Success, not yet received from majority

	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,3)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * Append response: from,term,success*/
	 
	sm := InitGenericSM()
	sm.InitLeader()
	//ev := MakeAppendEvent([]byte("xyz"))
	//sm.RouteEvent(ev)

	//sm.NextIndex[3]=2
	sm.CurrentTerm = 4
	sm.EntrySuccessCounter[int(sm.ThisLog.GetLastIndex())]=1
	sm.NextIndex[3]=int(sm.ThisLog.GetLastIndex())
	ev := MakeAppendEntriesRespEvent(3,4,true,int(sm.ThisLog.GetLastIndex()))
	sm.RouteEvent(ev)
	
	if(int64(sm.NextIndex[ev.From]) != sm.ThisLog.GetLastIndex() + 1){
		t.Errorf("Error1")
	}
		
	if(sm.EntrySuccessCounter[int(sm.ThisLog.GetLastIndex())] != 2) {
		t.Errorf("Error2")
	}
	os.RemoveAll("SMLogs")	
	
}

func TestLeaderAppendReq1(t *testing.T) {
	//Same term, No success

	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,3)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * */
	 
	sm := InitGenericSM()
	sm.InitLeader()

	var entriesToSend []*LogEntry
	
	ev := MakeAppendEntriesReqEvent(4,7,4,3,4,append(entriesToSend,&LogEntry{5,[]byte("abc")}),2)
	sm.RouteEvent(ev)
	
	if(sm.CurrentState != "follower") {
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")	
}

func TestLeaderAppendReq2(t *testing.T) {
	//Same term, No success

	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,3)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * */
	 
	sm := InitGenericSM()
	sm.InitLeader()

	var entriesToSend []*LogEntry
	
	ev := MakeAppendEntriesReqEvent(4,5,4,3,4,append(entriesToSend,&LogEntry{5,[]byte("abc")}),4)
	a := sm.RouteEvent(ev)
	ev1 := a[0].Send.EventToSend
	
	if(sm.CurrentState != "leader") {
		t.Errorf("Error1")
	}
	
	if(a[0].ActionType != "send" || a[0].Send.PeerId != ev.From) {
		t.Errorf("Error2")
	}
	
	if(ev1.AppendResp.Success != false) {
		t.Errorf("Error3")
	}
	os.RemoveAll("SMLogs")	
}

func TestLeaderTimeout(t *testing.T) {
	//Same term, No success

	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,3)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * */
	 
	sm := InitGenericSM()
	sm.InitLeader()
	
	ev := MakeTimeoutEvent()
	
	a := sm.RouteEvent(ev)
	
	if(len(a) != len(sm.KnownPeers) + 1){
		t.Errorf("Error2")
	}
	
	for i:=0;i<len(sm.KnownPeers);i++ {

		if(a[i].ActionType == "send"){
		
			ev1 := a[i].Send.EventToSend
		
		if(ev1.From != sm.MyId || ev1.Term != sm.CurrentTerm || ev1.AppendReq.PrevLogIndex != -1 || ev1.AppendReq.PrevLogTerm != -1){
			t.Errorf("Error4")
		}
	}
	}
	os.RemoveAll("SMLogs")	
}

func TestLeaderVoteReq1(t *testing.T){
	//VoteReq for higher term
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteReq: from int, term int, candidateId int, lastLogIndex int, lastLogTerm int*/
	 
	sm := InitGenericSM()
	sm.InitLeader()
	
	ev := MakeVoteReqEvent(4,7,3,4,4)

	sm.RouteEvent(ev)
	//voteReq for follower already tesed
	if(sm.CurrentState != "follower") {
		fmt.Println(sm.CurrentState)
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")	
}

func TestLeaderVoteReq2(t *testing.T){
	//VoteReq for lower term
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteReq: from int, term int, candidateId int, lastLogIndex int, lastLogTerm int*/
	 
	sm := InitGenericSM()
	sm.InitLeader()
	ev := MakeVoteReqEvent(4,5,4,4,4)

	a := sm.RouteEvent(ev)
	var a1 *Action
	for i:=0;i<len(a);i++ {
		if a[i].ActionType == "send" {
			a1 = a[i]
		}
	}	
	
	ev1 := a1.Send.EventToSend
	
	if(sm.CurrentState != "leader" || ev1.VoteResp.VoteGranted != false || a1.Send.PeerId != ev.From || ev1.Term != sm.CurrentTerm) {
		t.Errorf("Error")
	}
	os.RemoveAll("SMLogs")	
}
