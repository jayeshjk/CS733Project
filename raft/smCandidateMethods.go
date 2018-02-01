package raft

import (
	"math/rand"
//	"fmt"
//	"sync"
)

/*
 * 
 * */


func (sm *StateMachine) candidateAppendReq(ev *Event) ([]*Action){
	
	//1.Make empty slice of Action struct
	var arr []*Action
	
	//2.Extract data from Event object
	term := ev.Term
	
	//3.Logic
	sm.Lock()
	//defer sm.Unlock()
	if(term >= sm.CurrentTerm) {
		//Current or new leader discovered
		sm.CurrentState = "follower"
		sm.Unlock()
		arr = sm.followerAppendReq(ev)
		return arr
				
	} else {
		defer sm.Unlock()
		arr = append(arr, MakeSendAction(ev.From,MakeAppendEntriesRespEvent(sm.MyId,sm.CurrentTerm,false,ev.AppendReq.PrevLogIndex + 1)))
	}
	
	//sm.ActionChannel <- arr
	return arr
}

func (sm *StateMachine) candidateTimeout(ev *Event) ([]*Action) {
		
	var arr []*Action
	sm.Lock()
	defer sm.Unlock()

	majority := len(sm.KnownPeers)/2 + 1
		
	if(sm.NegativeVoteCounter > majority) {

	sm.CurrentState = "follower"
	arr = append(arr, MakeAlarmAction(ELECTION_TIMEOUT+rand.Intn(5)))

	} else if(sm.VoteCounter < majority){
			
		//exponential backoff might be needed here
		sm.CurrentState = "follower"
		arr = append(arr, MakeAlarmAction(FIX_PART + rand.Intn(10)))
		
	} else {
		
		sm.CurrentState = "leader"
		sm.KnownLeaderId = sm.MyId

		//Initialize leader data structures
		//sm.PrevSent = make([]int, len(sm.KnownPeers))
		//sm.NextIndex = make([]int, len(sm.KnownPeers))
		sm.NextIndex = make(map[int]int)
		sm.LogRepairing = make(map[int] bool)
			
		for i:=0;i<len(sm.KnownPeers);i++ {
			sm.NextIndex[sm.KnownPeers[i]] = /*sm.LastLogIndex*/ int(sm.ThisLog.GetLastIndex()) + 1
			sm.LogRepairing[sm.KnownPeers[i]] = false
		}
		sm.MatchIndex = make(map[int]int)
			
		for i:=0;i<len(sm.KnownPeers);i++ {
			sm.MatchIndex[i]= 0 
		}
			
		sm.EntrySuccessCounter = make(map[int]int)
		arr = append(arr, MakeAlarmAction(HEARTBEAT_TIMEOUT))
			
		i := 0
		totalKnownPeers := len(sm.KnownPeers)
		for ; i<totalKnownPeers; i++ {
		
			arr = append(arr,MakeSendAction(sm.KnownPeers[i],MakeHeartbeatEvent(sm.MyId,sm.CurrentTerm,sm.MyId,sm.CommitIndex))) //handle this in follower state
		
		}
	}
	//sm.ActionChannel <- arr
	return arr
}

func (sm *StateMachine) candidateVoteReq(ev *Event) ([]*Action){
	
	//fmt.Println("IN")
	//1.Make empty slice of Action struct
	var arr []*Action
	
	//2.Extract data from Event
	term := ev.Term
	candidateId := ev.VoteReq.CandidateId
	//lastLogIndex := ev.VoteReq.LastLogIndex
	//lastLogTerm := ev.VoteReq.LastLogTerm
	
	//3.Logic
	//fmt.Println("befor obtaining lock")
	sm.Lock()
	//fmt.Println("after obtaining lock")

	//defer sm.Unlock()

	if(term > sm.CurrentTerm) {
		
		//Candidate from old term
		sm.CurrentState = "follower"
		sm.VotedFor = -1
		//arr = append(arr, MakeAlarmAction(ELECTION_TIMEOUT))
		//arr = append(arr, MakeSendAction(candidateId, MakeVoteRespEvent(sm.MyId,sm.CurrentTerm,true)))
		sm.Unlock()
		arr = sm.followerVoteReq(ev)
		return arr
		
	} else {	
		defer sm.Unlock()
		arr = append(arr, MakeSendAction(candidateId, MakeVoteRespEvent(sm.MyId,sm.CurrentTerm,false)))
	} 
	
	//sm.ActionChannel <- arr
	return arr
}

func (sm StateMachine) candidateAppendResp(ev *Event) ([]*Action) {
	
	var arr []*Action
	arr = append(arr, MakeIgnoreAction())
	//sm.ActionChannel <- arr
	return arr
}

func (sm *StateMachine) candidateVoteResp(ev *Event) ([]*Action) {

	//fmt.Println("VoteResp was granted")

	//1. Array of action objects
	var arr []*Action
	
	//2.. Extract data from event object
	term := ev.Term
	voteGranted := ev.VoteResp.VoteGranted
	from := ev.From
	//3.Logic
	sm.Lock()
	defer sm.Unlock()

	if(term == sm.CurrentTerm && voteGranted == true) {
		//fmt.Println("Vote was granted for ",sm.MyId)
		if(sm.VoteReceived[from] == 0){
			
			sm.VoteReceived[from] = 1
			sm.VoteCounter++
			majority := len(sm.KnownPeers)/2 + 1
			if(sm.VoteCounter >= majority) {

				sm.CurrentState = "leader"
				sm.KnownLeaderId = sm.MyId
				//Initialize leader data structures
				sm.NextIndex = make(map[int]int)
				sm.LogRepairing = make(map[int] bool)
				
				for i:=0;i<len(sm.KnownPeers);i++ {
					sm.NextIndex[sm.KnownPeers[i]]=/*sm.LastLogIndex*/ int(sm.ThisLog.GetLastIndex()) + 1
					sm.LogRepairing[sm.KnownPeers[i]] = false
				}
				sm.MatchIndex = make(map[int]int)
			
				for i:=0;i<len(sm.KnownPeers);i++ {
					sm.MatchIndex[i]= 0 
				}
			
				sm.EntrySuccessCounter = make(map[int]int)
				arr = append(arr, MakeAlarmAction(HEARTBEAT_TIMEOUT))
				// send heatbeats
				i := 0
				totalKnownPeers := len(sm.KnownPeers)
				for ; i<totalKnownPeers; i++ {
		
					arr = append(arr,MakeSendAction(sm.KnownPeers[i],MakeHeartbeatEvent(sm.MyId,sm.CurrentTerm,sm.MyId,sm.CommitIndex))) //handle this in follower state
		
				}
				
			}
			
		} else {
			arr = append(arr, MakeIgnoreAction())
		}
		
	} else if(term > sm.CurrentTerm) {
		
		sm.CurrentState = "follower"
		sm.CurrentTerm = term
		sm.VotedFor = -1
		arr = append(arr, MakeAlarmAction(ELECTION_TIMEOUT+rand.Intn(5)))
		arr = append(arr, MakeWriteToFileAction("term",sm.CurrentTerm))
		arr = append(arr, MakeWriteToFileAction("vote",sm.VotedFor))

		
	} else if(term == sm.CurrentTerm && voteGranted == false){
		//fmt.Println("C",from)
		//fmt.Println("L",len(sm.VoteReceived))
		if(sm.VoteReceived[from] == 0){
			sm.VoteReceived[from]=-1
			sm.NegativeVoteCounter++
		
			majority := len(sm.KnownPeers)/2 + 1
			if(sm.NegativeVoteCounter > majority) {
	
				sm.CurrentState = "follower"
				arr = append(arr, MakeAlarmAction(ELECTION_TIMEOUT+rand.Intn(5)))

			}
		} else {
			arr = append(arr, MakeIgnoreAction())
		}

	}

	//sm.ActionChannel <- arr
	return arr
}
