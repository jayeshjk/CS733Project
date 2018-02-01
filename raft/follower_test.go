package raft

import(
	"testing"
	"fmt"
	"os"
)

/*
 * 
 * */

func TestFollowerAppendReq1(t *testing.T) {
	
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,3)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)*/
	 
	sm := InitGenericSM()

	//When term in message is lesser than currentTerm
	var entriesToSend []*LogEntry
	
	ev := MakeAppendEntriesReqEvent(4,3,4,3,4,append(entriesToSend,&LogEntry{5,[]byte("abc")}),2)
	a := sm.RouteEvent(ev)
	var a2 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "alarm"){
			//a1 = a[i]
		}else if(a[i].ActionType == "send"){
			a2 = a[i]
		}
	}
	
	a_type := a2.ActionType
	a_success := a2.Send.EventToSend.AppendResp.Success
	a_term := a2.Send.EventToSend.Term
	a_To := a2.Send.PeerId
	a_index	:= a2.Send.EventToSend.AppendResp.Index
	if(a_type != "send" || a_success != false || a_term != 5 || a_To != 4 || a_index != 4){
		t.Errorf("Error "+a_type)
	}
	//sm = nil
	os.RemoveAll("SMLogs")
	
}

func TestFollowerAppendReq2(t *testing.T) {
	//Successful append
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,3)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)*/
	 
	sm := InitGenericSM()
	var entriesToSend []*LogEntry

	ev := MakeAppendEntriesReqEvent(4,5,4,3,4,append(entriesToSend,&LogEntry{5,[]byte("abc")}),2)
	a := sm.RouteEvent(ev)
	var a1,a2 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "alarm"){
			a1 = a[i]
		}else if(a[i].ActionType == "send"){
			a2 = a[i]
		}
	}
	//a1 := a[0] //Alarm
	//a2 := a[1] //success response
	
	a1_type := a1.ActionType
	a2_type := a2.ActionType
	a2_event_type := a2.Send.EventToSend.EventType
	a2_success := a2.Send.EventToSend.AppendResp.Success
	a2_term := a2.Send.EventToSend.Term
	a2_To := a2.Send.PeerId
	a2_index := a2.Send.EventToSend.AppendResp.Index
		
	if(a1_type != "alarm" || a2_type != "send" || a2_event_type != "appendResp" || a2_success != true || a2_term != 5 || a2_To != 4 || a2_index != 4){
		t.Errorf("Error "+a1_type)
	} 
	//sm = nil
	os.RemoveAll("SMLogs")
	
}

func TestFollowerAppendReq3(t *testing.T) {
	//PrevLogTerm different
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * AppendReq: from,term,leaderId,prevLogIndex,prevLogTerm,entries,
	 * leaderCommit*/
	 
	sm := InitGenericSM()
	var entriesToSend []*LogEntry

	ev := MakeAppendEntriesReqEvent(4,5,4,2,3,append(entriesToSend,&LogEntry{5,[]byte("abc")}),2)
	a := sm.RouteEvent(ev)
	
	var a1,a2 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "alarm"){
			a1 = a[i]
		}else if(a[i].ActionType == "send"){
			a2 = a[i]
		}
	}	
	a1_type := a1.ActionType
	a2_type := a2.ActionType
	a2_event_type := a2.Send.EventToSend.EventType
	a2_success := a2.Send.EventToSend.AppendResp.Success
	a2_term := a2.Send.EventToSend.Term
	a2_To := a2.Send.PeerId
	a2_index := a2.Send.EventToSend.AppendResp.Index
		
	if(a1_type != "alarm" || a2_type != "send" || a2_event_type != "appendResp" || a2_success != false || a2_term != 5 || a2_To != ev.From || a2_index != 3){
		t.Errorf("Error "+a1_type)
	} 
	//sm = nil
	os.RemoveAll("SMLogs")
	
}

func TestFollowerAppendReq4(t *testing.T) {
	//Successful append to log
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * AppendReq: from,term,leaderId,prevLogIndex,prevLogTerm,entries,
	 * leaderCommit*/
	 
	sm := InitGenericSM()
	var entriesToSend []*LogEntry
	ev := MakeAppendEntriesReqEvent(4,5,4,2,2,append(entriesToSend,&LogEntry{5,[]byte("xyz")}),2)

	a := sm.RouteEvent(ev)
	
	var a1,a2 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "alarm"){
			a1 = a[i]
		}else if(a[i].ActionType == "send"){
			a2 = a[i]
		}
	}
	//a1 := a[0] //Alarm
	//a2 := a[1] //success response
	
	a1_type := a1.ActionType
	a2_type := a2.ActionType
	a2_event_type := a2.Send.EventToSend.EventType
	a2_success := a2.Send.EventToSend.AppendResp.Success
	a2_index := a2.Send.EventToSend.AppendResp.Index
	a2_term := a2.Send.EventToSend.Term
	a2_To := a2.Send.PeerId
	

	if(a1_type != "alarm" || a2_type != "send" || a2_event_type != "appendResp" || a2_success != true || a2_term != 5 || a2_To != 4 || a2_index != 3){
		t.Errorf("Error "+a1_type)
	} 
	os.RemoveAll("SMLogs")
	
}

func TestFollowerAppendReq5(t *testing.T) {
	//PrevLogIndex greater than sm.LastLogIndex
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * AppendReq: from,term,leaderId,prevLogIndex,prevLogTerm,entries,
	 * leaderCommit*/
	 
	sm := InitGenericSM()
	var entriesToSend []*LogEntry
	//fmt.Println(sm)
	ev := MakeAppendEntriesReqEvent(4,5,4,5,5,append(entriesToSend,&LogEntry{5,[]byte("xyz")}),2)

	//sm_ptr := &sm
	a := sm.RouteEvent(ev)
	
	var a1,a2 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "alarm"){
			a1 = a[i]
		}else if(a[i].ActionType == "send"){
			a2 = a[i]
		}
	}
	//a1 := a[0] //Alarm
	//a2 := a[1] //failure response
	
	a1_type := a1.ActionType
	a2_type := a2.ActionType
	a2_event_type := a2.Send.EventToSend.EventType
	a2_success := a2.Send.EventToSend.AppendResp.Success
	a2_term := a2.Send.EventToSend.Term
	a2_To := a2.Send.PeerId
	a2_index := a2.Send.EventToSend.AppendResp.Index
	
	if(a1_type != "alarm" || a2_type != "send" || a2_To != 4 || a2_event_type != "appendResp" || a2_success != false || a2_term != 5 || a2_index != 6){
		t.Errorf("Error "+a1_type)
	}
	os.RemoveAll("SMLogs")
	 
}

func TestFollowerTimeout(t *testing.T) {
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * */
	 
	sm := InitGenericSM()
	ev := MakeTimeoutEvent()
	
	a := sm.RouteEvent(ev)
	
	if(sm.CurrentState != "candidate") {
		t.Errorf("Error ")
	}
	
	
	for i := 0;i<len(a);i++ {
		/*if(a[i].ActionType != "send" || a[i].Send.PeerId != sm.KnownPeers[i]){
			t.Errorf("Error ")
		}*/
		if(a[i].ActionType == "send"){
			
			ev1 := a[i].Send.EventToSend.VoteReq
		
			if(a[i].Send.EventToSend.Term != 6 || ev1.CandidateId != 1 || ev1.LastLogIndex != 3 || ev1.LastLogTerm != 4) {
				fmt.Println(a[i].Send.EventToSend.Term)
				t.Errorf("Error ")
			}
		}
	}
	os.RemoveAll("SMLogs")
	
}

func TestFollowerVoteReq1(t *testing.T) {
	
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteReq parameters: from, term, candidateId, lastLogIndex, lastLogTerm
	 * */
	 
	sm := InitGenericSM()
	sm.VotedFor = 3 //Same voteReq received again, maybe due to network failure
	ev := MakeVoteReqEvent(3,6,3,5,5)
	
	a := sm.RouteEvent(ev)
	var a1 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "send"){
			a1 = a[i]
		}
	}
	ev1 := a1.Send.EventToSend.VoteResp
	
	if(a1.ActionType != "send" || a1.Send.PeerId != ev.From || a1.Send.EventToSend.Term != ev.Term || sm.CurrentTerm != ev.Term) {
		
		t.Errorf("Error1")
	} 
	if(ev1.VoteGranted != true){
		t.Errorf("Error2")
	}
	os.RemoveAll("SMLogs")
}

func TestFollowerVoteReq2(t *testing.T) {
	//VoteReq for older term
	
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteReq parameters: from, term, candidateId, lastLogIndex, lastLogTerm
	 * */
	 
	sm := InitGenericSM()
	sm.VotedFor = -1
	ev := MakeVoteReqEvent(3,2,3,5,5)
	
	a := sm.RouteEvent(ev)
	
	var a1 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "send"){
			a1 = a[i]
		}
	}
	ev1 := a1.Send.EventToSend.VoteResp
	
	if(a1.ActionType != "send" || a1.Send.PeerId != ev.From || a1.Send.EventToSend.Term != sm.CurrentTerm ) {
		
		t.Errorf("Error1",a1.Send.EventToSend.Term)
	} 
	if(ev1.VoteGranted != false){
		t.Errorf("Error2")
	}
	os.RemoveAll("SMLogs")
}

func TestFollowerVoteReq3(t *testing.T) {
	//VoteReq with higher term but older logTerm at the end
	
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteReq parameters: from, term, candidateId, lastLogIndex, lastLogTerm
	 * */
	 
	sm := InitGenericSM()
	sm.VotedFor = -1
	ev := MakeVoteReqEvent(3,7,3,2,2)
	
	a := sm.RouteEvent(ev)
	
	var a1 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "send"){
			a1 = a[i]
		}
	}
	ev1 := a1.Send.EventToSend.VoteResp
	
	if(a1.ActionType != "send" || a1.Send.PeerId != ev.From || a1.Send.EventToSend.Term != sm.CurrentTerm ) {
		
		t.Errorf("Error1")
	} 
	if(ev1.VoteGranted != false){
		t.Errorf("Error2")
	}
	os.RemoveAll("SMLogs")
}

func TestFollowerVoteReq4(t *testing.T) {
	//VoteReq with higher term but shorter log and same logTerm at the end
	
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteReq parameters: from, term, candidateId, lastLogIndex, lastLogTerm
	 * */
	 
	sm := InitGenericSM()
	sm.VotedFor = -1
	ev := MakeVoteReqEvent(3,7,3,2,4)
	
	a := sm.RouteEvent(ev)
	
	var a1 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "send"){
			a1 = a[i]
		}
	}
	ev1 := a1.Send.EventToSend.VoteResp
	
	if(a1.ActionType != "send" || a1.Send.PeerId != ev.From || a1.Send.EventToSend.Term != sm.CurrentTerm) {
		
			t.Errorf("Error1")
	}
	 
	if(ev1.VoteGranted != false){
		t.Errorf("Error2")
	}
	os.RemoveAll("SMLogs")
}

func TestFollowerVoteReq5(t *testing.T) {
	//VoteReq with higher term but shorter log and higher Term at the end
	
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteReq parameters: from, term, candidateId, lastLogIndex, lastLogTerm
	 * */
	 
	sm := InitGenericSM()
	sm.VotedFor = -1
	ev := MakeVoteReqEvent(3,7,3,2,5)
	
	a := sm.RouteEvent(ev)
	
	var a1 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "send"){
			a1 = a[i]
		}
	}
	ev1 := a1.Send.EventToSend.VoteResp
	
	if(a1.ActionType != "send" || a1.Send.PeerId != ev.From || a1.Send.EventToSend.Term != ev.Term || sm.CurrentTerm != ev.Term) {
		
		t.Errorf("Error1")
	} 
	if(ev1.VoteGranted != true){
		t.Errorf("Error2")
	}
	os.RemoveAll("SMLogs")
}

func TestFollowerVoteReq6(t *testing.T) {
	//Follower already voted
	
	/*State Machine Initialization.
	 id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,4)(these are terms.)
	 CommitIndex=2, LastLogIndex=3, KnownPeers=(1,2,3,4,5)
	 * VoteReq parameters: from, term, candidateId, lastLogIndex, lastLogTerm
	 * */
	 
	sm := InitGenericSM()
	sm.VotedFor = 2
	ev := MakeVoteReqEvent(3,5,3,2,5)
	
	a := sm.RouteEvent(ev)
	
	var a1 *Action
	for i:=0;i<len(a);i++ {
		if(a[i].ActionType == "send"){
			a1 = a[i]
		}
	}
	ev1 := a1.Send.EventToSend.VoteResp
	
	if(a1.ActionType != "send" || a1.Send.PeerId != ev.From || a1.Send.EventToSend.Term != sm.CurrentTerm ) {
		
		t.Errorf("Error1")
	} 
	if(ev1.VoteGranted != false){
		fmt.Println()
		t.Errorf("Error2")
	}
	os.RemoveAll("SMLogs")
}

