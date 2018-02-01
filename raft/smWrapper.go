package raft

import (
	"fmt"
	"github.com/cs733-iitb/log"
	"strconv"	
	"os"
)

/*
 * 
 * 
 * */
 
 func MakeSM() *StateMachine {
	 sm := StateMachine{CurrentState:"follower",CurrentTerm:0,VotedFor:-1}
	 //sm.Log.Entries = make(map[int]*LogEntry)

	 return &sm
 }
 
 /*
  * This function was written for testing purposes.
  * Returns a state machine with following parameters. Make changes
  * in this state machine as required.
  * id:1, CurrentTerm: 5, VotedFor:4, Log=(1,1,2,3)(these are terms.)
  * CommitIndex=2, LastLogIndex=3, KnownPeers=(2,3,4,5,6)
  * */
 func InitGenericSM() *StateMachine {
	
	os.RemoveAll("SMLogs")
	sm := MakeSM()
	sm.CurrentTerm = 5
	sm.VotedFor = 4
	sm.MyId = 1
	
	tempLogName := "SMLogs/LogOf"+strconv.Itoa(sm.MyId)
	tempLog, err := log.Open(tempLogName)
	if(err != nil){
		fmt.Println("Error while opening the Log: "+ tempLogName)
	}else {
		sm.ThisLog = tempLog
		sm.ThisLog.RegisterSampleEntry(LogEntry{})
	}
	
	//entries map[int]*LogEntry
	entries := make(map[int]*LogEntry)
	
	entries[0] = &LogEntry{Term:1,Data:[]byte("abc")}
	sm.ThisLog.Append(*(entries[0]))
	entries[1] = &LogEntry{Term:1,Data:[]byte("abc")}
	sm.ThisLog.Append(*(entries[1]))
	entries[2] = &LogEntry{Term:2,Data:[]byte("abc")}
	sm.ThisLog.Append(*(entries[2]))	
	entries[3] = &LogEntry{Term:4,Data:[]byte("abc")}
	sm.ThisLog.Append(*(entries[3]))
	//sm.Log = LogStruct {Entries:entries}
	
	sm.CommitIndex = 2
	//sm.LastLogIndex = len(sm.Log.Entries) - 1
	
	sm.KnownPeers=append(sm.KnownPeers,2)
	sm.KnownPeers=append(sm.KnownPeers,3)
	sm.KnownPeers=append(sm.KnownPeers,4)	
	sm.KnownPeers=append(sm.KnownPeers,5) 
	sm.KnownPeers=append(sm.KnownPeers,6)
	
	sm.VoteReceived = make(map[int]int)
	for i:=0 ;i<len(sm.KnownPeers);i++ {
		sm.VoteReceived[sm.KnownPeers[i]]=0
	}
	return sm;
 }
 
 func (sm *StateMachine) InitLeader() {
	sm.CurrentState = "leader"
	//Initialize leader data structures
	sm.NextIndex = make(map[int]int)
	sm.LogRepairing = make(map[int] bool)		
	for i:=0;i<len(sm.KnownPeers);i++ {
		sm.NextIndex[sm.KnownPeers[i]]=int(sm.ThisLog.GetLastIndex()) + 1
	}
	sm.MatchIndex = make(map[int]int)
			
	for i:=0;i<len(sm.KnownPeers);i++ {
		sm.MatchIndex[i]= 0 
	}
			
	sm.EntrySuccessCounter = make(map[int]int)

 }
 
 
 func (sm StateMachine) SetCurrentState(state string){

		sm.CurrentState=state

 }
 
  func (sm StateMachine) SetCurrentTerm(term int){

	 sm.CurrentTerm=term

 }
 
  func (sm StateMachine) SetId(id int){
		
		sm.MyId=id
		
 }
 
  func (sm StateMachine) SetVotedFor(id int){
	  
	  sm.VotedFor=id
	 
 }
 
  func (sm StateMachine) SetCommitIndex(index int){
	 
	 sm.CommitIndex=index
 }
 
/*
 * Possible events: append,appendReq,appendResp,timeout,voteReq,voteResp
 * */
func (sm *StateMachine) RouteEvent(ev *Event) ([]*Action){
	
	//fmt.Println("In the event Route")
	var arr []*Action
	
	if(sm.CurrentState=="follower") {
		if(ev.EventType=="append") {
			//1.Create corresponding array of action objects
			//a := MakeCommitAction(0,nil,true)
			//arr := [] *Action{a} 
			
			arr = append(arr,MakeCommitAction(0,nil,true))
			arr = append(arr,MakeCommitAction(0,nil,true))
			//2. Put it in action channel
			//fmt.Println("Action put in channel")
			//sm.ActionChannel <- arr
		
		} else if(ev.EventType=="appendReq") {
			
			arr=sm.followerAppendReq(ev)
			
		} else if(ev.EventType=="appendResp") {
			
			arr=sm.followerAppendResp(ev)
			
		} else if(ev.EventType=="timeout") {
			
			arr=sm.followerTimeout(ev)
					
		} else if(ev.EventType=="voteReq") {
			
			arr=sm.followerVoteReq(ev)
			
		} else if(ev.EventType=="voteResp") {
			arr=sm.followerVoteResp(ev)
		}
	
	}else if(sm.CurrentState == "candidate") {
		if(ev.EventType=="append") {
			//1.Create corresponding array of action objects
			//a := MakeCommitAction(0,nil,true)
			//arr = [] *Action{a} 
			arr = append(arr,MakeCommitAction(0,nil,true))
			arr = append(arr,MakeCommitAction(0,nil,true))
			//2. Put it in action channel
			//fmt.Println("Action put in channel")
			//sm.ActionChannel <- arr
		
		} else if(ev.EventType=="appendReq") {
			
			arr=sm.candidateAppendReq(ev)
			
		} else if(ev.EventType=="appendResp") {
			
			arr=sm.candidateAppendResp(ev)
			
		} else if(ev.EventType=="timeout") {
			
			arr=sm.candidateTimeout(ev)
					
		} else if(ev.EventType=="voteReq") {
			
			arr=sm.candidateVoteReq(ev)
			
		} else if(ev.EventType=="voteResp") {
			
			arr=sm.candidateVoteResp(ev)
		
		}
	} else if(sm.CurrentState == "leader") {
		if(ev.EventType=="append") {
			
			arr=sm.leaderAppend(ev)
			
		} else if(ev.EventType=="appendReq") {
			
			arr=sm.leaderAppendReq(ev)
			
		} else if(ev.EventType=="appendResp") {
			
			arr=sm.leaderAppendResp(ev)
			
		} else if(ev.EventType=="timeout") {
			
			arr=sm.leaderTimeout(ev)
					
		} else if(ev.EventType=="voteReq") {
			
			arr=sm.leaderVoteReq(ev)
			
		} else if(ev.EventType=="voteResp") {
			
			arr=sm.leaderVoteResp(ev)
		
		}
	}
	
	//sm.ActionChannel <- arr
	
	return arr;
}
