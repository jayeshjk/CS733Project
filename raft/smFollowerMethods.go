package raft

import(
	"fmt"
	"math/rand"
//	"sync"

)

/*
 * 
 * */
 
func (sm *StateMachine) followerAppendReq(ev *Event) ([]*Action) {
	
	//1.Make empty slice of Action struct
	var arr []*Action
	
	//2.Extract data from Event object
	term := ev.Term
	//leaderId := ev.AppendReq.LeaderId
	prevLogIndex := ev.AppendReq.PrevLogIndex
	prevLogTerm := ev.AppendReq.PrevLogTerm
	entries := ev.AppendReq.Entries
	leaderCommit := ev.AppendReq.LeaderCommit
	if(len(entries) > 0){
		//fmt.Println()
	}//Get info about last entry in the log
	sm.Lock()
	defer sm.Unlock()
	
	persistentLastLogIndex := sm.ThisLog.GetLastIndex()
	persistentPrevTerm := -1
	if(prevLogIndex != -1 && int64(prevLogIndex) <= persistentLastLogIndex) {
		temp,err := sm.ThisLog.Get(int64(prevLogIndex))
		if(err != nil){
			fmt.Println("Error",err.Error())
		}else{
			persistentPrevTerm = temp.(LogEntry).Term
		}
	} else{
		persistentPrevTerm = -1
	}
	
	//3.Logic
	//TODO Handle heartbeats
	if(term < sm.CurrentTerm) {
		//Wrong Leader
		arr = append(arr, MakeSendAction(ev.From,MakeAppendEntriesRespEvent(sm.MyId,sm.CurrentTerm,false,prevLogIndex+1)))
		
	} else {
		//Correcct Leader
		sm.KnownLeaderId = ev.From
		if(entries != nil) {
			//fmt.Println(sm.MyId," append data ",ev.AppendReq)
		}

		arr = append(arr, MakeAlarmAction(FIX_PART+ELECTION_TIMEOUT+rand.Intn(5)))
			
		if(term > sm.CurrentTerm){
			
			 sm.CurrentTerm = term
			 sm.VotedFor = -1
			 arr = append(arr, MakeWriteToFileAction("term",sm.CurrentTerm))
			 arr = append(arr, MakeWriteToFileAction("vote",sm.VotedFor))

		}
		 
		if(entries == nil){
			//Heartbeat message
			//alarm already set above
			//do the commit processing
			if(leaderCommit >= sm.CommitIndex) {
								//fmt.Println("",leaderCommit,"+",sm.CommitIndex)

				for i:= sm.CommitIndex+1;i<=leaderCommit && int64(i) <= sm.ThisLog.GetLastIndex();i++{
					
					tempCommit,_ := sm.ThisLog.Get(int64(i))
					tempLogEntry := tempCommit.(LogEntry)
					arr=append(arr, MakeCommitAction(i,tempLogEntry.Data,false))
					//arr = append(arr,MakeApplyAction(i))
					sm.CommitIndex++
				}
				//sm.CommitIndex=leaderCommit
				//TODO: Remove unncessary entries from SM log(entries already commmited)
			}
			
		} else if(int64(prevLogIndex) > persistentLastLogIndex) {
			
			arr = append(arr, MakeSendAction(ev.From,MakeAppendEntriesRespEvent(sm.MyId,sm.CurrentTerm,false,prevLogIndex+1)))
			//fmt.Println(sm.MyId," append data ",ev.AppendReq)
		
		}else if(prevLogIndex != -1 && /*sm.Log.Entries[prevLogIndex].Term*/ persistentPrevTerm != prevLogTerm) {
			//check the term at previous log index term
			
			arr = append(arr, MakeSendAction(ev.From,MakeAppendEntriesRespEvent(sm.MyId,sm.CurrentTerm,false,prevLogIndex+1)))
				
		}else {

			//Previous index term matches
			//fmt.Println("Gonna append to the log")
			//fmt.Println(len(sm.Log.Entries))
			if(/*sm.LastLogIndex*/ persistentLastLogIndex >= int64(prevLogIndex) + 1) {
				arr = append(arr,MakeLogTruncateAction(prevLogIndex + 1))
			
			}
			//First parameter is a dummy
			arr = append(arr, MakeLogStoreAction(-1,entries[0]))
			
			arr = append(arr, MakeSendAction(ev.From,MakeAppendEntriesRespEvent(sm.MyId,sm.CurrentTerm,true,prevLogIndex+1)))
				
			if(leaderCommit >= sm.CommitIndex) {
				for i:= sm.CommitIndex+1;i<=leaderCommit && int64(i) <= sm.ThisLog.GetLastIndex();i++{
					
					tempCommit,_ := sm.ThisLog.Get(int64(i))
					tempLogEntry := tempCommit.(LogEntry)
					arr=append(arr, MakeCommitAction(i,tempLogEntry.Data,false))
					//arr = append(arr,MakeApplyAction(i))
					sm.CommitIndex++
					
				}
				//sm.CommitIndex=leaderCommit
				
				
				//TODO: Remove unncessary entries from SM log(entries already commmited)
			}
			
		}
	}
	
	//4.Send on action channel
	//sm.ActionChannel <- arr
	return arr
}

func (sm *StateMachine) followerTimeout(ev *Event) ([]*Action) {
	
	sm.Lock()
	defer sm.Unlock()
	//1.Incr Current Term
	sm.CurrentTerm++
	var arr []*Action
	
	//2. Send VoteReq to all peers
	sm.CurrentState = "candidate"
	//fmt.Println("Transistion",sm.MyId)
	sm.VoteCounter = 0
	sm.NegativeVoteCounter=0
	sm.VotedFor = sm.MyId
	sm.VoteReceived = make(map[int]int)
	arr = append(arr, MakeAlarmAction(ELECTION_TIMEOUT+rand.Intn(5)+CANDIDATE_TIMEOUT))

	arr = append(arr, MakeWriteToFileAction("term",sm.CurrentTerm))
	arr = append(arr, MakeWriteToFileAction("vote",sm.VotedFor))

	i := 0
	totalKnownPeers := len(sm.KnownPeers)
	persistentLastLogIndex := sm.ThisLog.GetLastIndex()

	for ; i<totalKnownPeers; i++ {
		
		if(/*sm.LastLogIndex*/ persistentLastLogIndex == -1){
			arr = append(arr,MakeSendAction(sm.KnownPeers[i],MakeVoteReqEvent(sm.MyId,sm.CurrentTerm,sm.MyId,-1,-1)))
		}else{
			temp,_ := sm.ThisLog.Get(persistentLastLogIndex)
			persistentLastLogTerm := temp.(LogEntry).Term
			arr = append(arr,MakeSendAction(sm.KnownPeers[i],MakeVoteReqEvent(sm.MyId,sm.CurrentTerm,sm.MyId,int(persistentLastLogIndex),persistentLastLogTerm)))
		}
		//sm.VoteReceived=append(sm.VoteReceived,0)
		sm.VoteReceived[sm.KnownPeers[i]]=0
	}
	
	//3.Set Election over timeout
	
	//4.Transition to candidate state
	//5. Send over Action channel
	//sm.ActionChannel <- arr
	
	return arr
}

func (sm *StateMachine) followerVoteReq(ev *Event) ([]*Action){
	
	//1.Make empty slice of Action struct
	var arr []*Action
	
	//2.Extract data from Event
	term := ev.Term
	candidateId := ev.VoteReq.CandidateId
	lastLogIndex := ev.VoteReq.LastLogIndex
	lastLogTerm := ev.VoteReq.LastLogTerm
	

	//Get persistent info from log
	persistentLastLogTerm := -1
	persistentLastLogIndex := sm.ThisLog.GetLastIndex()
	if(persistentLastLogIndex != -1) {
		temp,_ := sm.ThisLog.Get(persistentLastLogIndex)
		persistentLastLogTerm = temp.(LogEntry).Term
	} else{
		persistentLastLogTerm = -1
	}

	
	//3.Logic
	sm.Lock()
	defer sm.Unlock()
	if(term > sm.CurrentTerm){
		
		sm.CurrentTerm = term
		sm.VotedFor = -1
		arr = append(arr, MakeWriteToFileAction("term",sm.CurrentTerm))
		arr = append(arr, MakeWriteToFileAction("vote",sm.VotedFor))
	}
	if(term < sm.CurrentTerm) {
			
		//Candidate from old term
		arr = append(arr, MakeSendAction(candidateId, MakeVoteRespEvent(sm.MyId,sm.CurrentTerm,false)))
	
	} else if(sm.VotedFor!=-1 && sm.VotedFor != candidateId){
	
		//Has this follower already voted?
		arr = append(arr, MakeSendAction(candidateId, MakeVoteRespEvent(sm.MyId,sm.CurrentTerm,false)))
		//fmt.Println("VoteReq received at ",sm.MyId," from ",candidateId)
		//fmt.Println("SM Term: ",sm.CurrentTerm," Message Term:",term," VotedFor:",sm.VotedFor)
	
	}else if(/*sm.LastLogIndex*/ persistentLastLogIndex == -1){
		//Log is empty, means there is no point in checking further. Current Term is fine, follower has not voted, so vote now.
		sm.CurrentTerm = term
		sm.VotedFor = candidateId
		arr = append(arr, MakeAlarmAction(FIX_PART+ELECTION_TIMEOUT+rand.Intn(5)))

		arr = append(arr, MakeWriteToFileAction("term",sm.CurrentTerm))
		arr = append(arr, MakeWriteToFileAction("vote",sm.VotedFor))

		arr = append(arr, MakeSendAction(candidateId, MakeVoteRespEvent(sm.MyId,sm.CurrentTerm,true)))
		
	}else if(lastLogTerm < persistentLastLogTerm/*sm.Log.Entries[sm.LastLogIndex].Term*/) {

		//First check the term of last entries of both this follower and candidate
		arr = append(arr, MakeSendAction(candidateId, MakeVoteRespEvent(sm.MyId,sm.CurrentTerm,false)))

	} else if(lastLogTerm == persistentLastLogTerm /*sm.Log.Entries[sm.LastLogIndex].Term*/ && int64(lastLogIndex) < persistentLastLogIndex/*sm.LastLogIndex*/) {
		
		//Then, if terms are same, check whole log is longer
		arr = append(arr, MakeSendAction(candidateId, MakeVoteRespEvent(sm.MyId,sm.CurrentTerm,false)))

	} else {
		
		//Now, you have not voted and either term is greater than this follower's term or candidate's log is longer
		//so vote for this candidate
		sm.CurrentTerm = term
		sm.VotedFor = candidateId
		arr = append(arr, MakeAlarmAction(FIX_PART+ELECTION_TIMEOUT+rand.Intn(5)))

		arr = append(arr, MakeWriteToFileAction("term",sm.CurrentTerm))
		arr = append(arr, MakeWriteToFileAction("vote",sm.VotedFor))

		arr = append(arr, MakeSendAction(candidateId, MakeVoteRespEvent(sm.MyId,sm.CurrentTerm,true)))

	}
	
	//sm.ActionChannel <- arr
	return arr
}

func (sm *StateMachine) followerAppendResp(ev *Event) ([]*Action){
	
	var arr []*Action
	arr = append(arr, MakeIgnoreAction())
	//sm.ActionChannel <- arr
	return arr
}

func (sm *StateMachine) followerVoteResp(ev *Event) ([]*Action){

	var arr []*Action
	arr = append(arr, MakeIgnoreAction())
	//sm.ActionChannel <- arr
	return arr
}
