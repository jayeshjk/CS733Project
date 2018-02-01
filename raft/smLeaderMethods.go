package raft

import(
	"fmt"
	"math/rand"
//	"sync"
)

/*
 * 
 * 
 * */

func (sm *StateMachine) leaderAppend(ev *Event) ([]*Action) {

	//1.Make empty slice of Action struct
	var arr []*Action
	
	//2.Exttract the data
	data := ev.Append.Data

	arr=append(arr,MakeAlarmAction(HEARTBEAT_TIMEOUT))	
	//3. Append to your own log

	sm.Lock()
	defer sm.Unlock()

	map_mutex.Lock()
	sm.EntrySuccessCounter[int(sm.ThisLog.GetLastIndex()) + 1] = 1
	//fmt.Println(sm.EntrySuccessCounter)
	map_mutex.Unlock()
	tempEntry := &LogEntry{Term:sm.CurrentTerm , Data:data}
	arr = append(arr,MakeLogStoreAction(int(sm.ThisLog.GetLastIndex()) + 1,tempEntry/*sm.Log.Entries[sm.LastLogIndex]*/))
	var entriesToSend []*LogEntry
	//4. Send this to all KnownPeers
	totalPeers := len(sm.KnownPeers)
	i:=0
	for ;i<totalPeers;i++ {
					
		if(int64(sm.NextIndex[sm.KnownPeers[i]]) == sm.ThisLog.GetLastIndex() + 1 || sm.LogRepairing[sm.KnownPeers[i]] == false) {
			
			sm.LogRepairing[sm.KnownPeers[i]] = false
			sm.NextIndex[sm.KnownPeers[i]] = int(sm.ThisLog.GetLastIndex()) + 1
			//fmt.Println(sm.LastLogIndex-1)
			if(sm.ThisLog.GetLastIndex() == -1){
				arr=append(arr,MakeSendAction(sm.KnownPeers[i],MakeAppendEntriesReqEvent(sm.MyId,sm.CurrentTerm,sm.MyId,-1,-1,append(entriesToSend,tempEntry),sm.CommitIndex)))

			} else{
				
				persistentLastLogIndex := sm.ThisLog.GetLastIndex()
				temp,_ := sm.ThisLog.Get(persistentLastLogIndex)
				persistentLastLogTerm := temp.(LogEntry).Term
				arr=append(arr,MakeSendAction(sm.KnownPeers[i],MakeAppendEntriesReqEvent(sm.MyId,sm.CurrentTerm,sm.MyId,int(persistentLastLogIndex),persistentLastLogTerm,append(entriesToSend,tempEntry),sm.CommitIndex)))
	
			}	
		} 
	}
	
	return arr
}

func (sm *StateMachine) leaderAppendReq(ev *Event) ([]*Action) {
	
	//1.Make empty slice of Action struct
	var arr []*Action
	
	//2.Extract data from Event object
	term := ev.Term
	
	//3.Logic
	sm.Lock()
	//defer sm.Unlock()

	if(term > sm.CurrentTerm) {
		//Become a follower
		sm.CurrentState = "follower"
		sm.Unlock()
		arr = sm.followerAppendReq(ev)
		return arr
	} else {
		defer sm.Unlock()	
		arr = append(arr, MakeSendAction(ev.From,MakeAppendEntriesRespEvent(sm.MyId,sm.CurrentTerm,false,ev.AppendReq.PrevLogIndex + 1)))	

	}
	
	//4.Send on action channel
	//sm.ActionChannel <- arr
	return arr
}

func (sm *StateMachine) leaderTimeout(ev *Event) ([]*Action) {
	
	//1.Incr Current Term
	var arr []*Action
	
	//3.Reset heartbeat timeout
	arr = append(arr, MakeAlarmAction(HEARTBEAT_TIMEOUT))

	//2. Send AppendReq/heartbeat to all peers
	sm.Lock()
	defer sm.Unlock()

	i := 0
	totalKnownPeers := len(sm.KnownPeers)
	for ; i<totalKnownPeers; i++ {
		
		arr = append(arr,MakeSendAction(sm.KnownPeers[i],MakeHeartbeatEvent(sm.MyId,sm.CurrentTerm,sm.MyId,sm.CommitIndex))) //handle this in follower state
		
	}
		
	//sm.ActionChannel <- arr
	return arr
}

func (sm *StateMachine) leaderVoteReq(ev *Event) ([]*Action){
	
	//1.Make empty slice of Action struct
	var arr []*Action
	
	term := ev.Term
	candidateId := ev.VoteReq.CandidateId

	sm.Lock()
	//defer sm.Unlock()

	if(term <= sm.CurrentTerm) {
		defer sm.Unlock()
		arr = append(arr, MakeSendAction(candidateId, MakeVoteRespEvent(sm.MyId,sm.CurrentTerm,false)))

	} else {
		//Once election starts, it has to continue
		sm.CurrentState = "follower"
		sm.Unlock()
		arr = sm.followerVoteReq(ev)
		return arr
	}

	//sm.ActionChannel <- arr	
	return arr
}

func (sm *StateMachine) leaderAppendResp(ev *Event) ([]*Action) {
	
	//1. Array of action objects
	var arr []*Action
	
	//2. Extract data from Action objects
	from := ev.From
	term := ev.Term
	success := ev.AppendResp.Success
	lastIndex := ev.AppendResp.Index
	
	//3.Logic
	sm.Lock()
	defer sm.Unlock()
	
	if(term > sm.CurrentTerm) {
		sm.CurrentState = "follower"
		sm.CurrentTerm = term
		sm.VotedFor = -1
		
		arr = append(arr, MakeAlarmAction(ELECTION_TIMEOUT+rand.Intn(5)))
		arr = append(arr, MakeWriteToFileAction("term",sm.CurrentTerm))
		arr = append(arr, MakeWriteToFileAction("vote",sm.VotedFor))
	
	} else if(term==sm.CurrentTerm && success == false){
		
		/*arr = append(arr,MakeSendAction(from,MakeAppendEntriesReqEvent(sm.MyId,sm.CurrentTerm,sm.MyId,sm.PrevSent[from]-2,sm.Log.Entries[sm.PrevSent[from]-2].Term,sm.Log.Entries[sm.PrevSent[from]-1].Data,sm.CommitIndex)))
		
			
		sm.PrevSent[from]--*/
		var entriesToSend []*LogEntry
		if(lastIndex == sm.NextIndex[from]){
			
			sm.NextIndex[from]--;
			//Add a check for negative index here.
		
			sm.LogRepairing[from] = true
			if(sm.NextIndex[from]==0){
			
			//persistentLastLogIndex := sm.ThisLog.GetLastIndex()
				temp,_ := sm.ThisLog.Get(0)
				persistentLogEntry := temp.(LogEntry)
				arr=append(arr,MakeSendAction(from,MakeAppendEntriesReqEvent(sm.MyId,sm.CurrentTerm,sm.MyId,-1,-1,append(entriesToSend,&persistentLogEntry),sm.CommitIndex)))
			}else{
				temp1,_ := sm.ThisLog.Get(int64(sm.NextIndex[from])-1)
				persistentPrevLogEntry := temp1.(LogEntry)
				temp,_ := sm.ThisLog.Get(int64(sm.NextIndex[from]))
				persistentLogEntryToSend := temp.(LogEntry)
			
				arr=append(arr,MakeSendAction(from,MakeAppendEntriesReqEvent(sm.MyId,sm.CurrentTerm,sm.MyId,sm.NextIndex[from]-1,persistentPrevLogEntry.Term,append(entriesToSend,&persistentLogEntryToSend),sm.CommitIndex)))
			}
		}
	
	} else if(term == sm.CurrentTerm && success == true) {
		//fmt.Println("IN leader received success for:",lastIndex," ",sm.NextIndex[from])
		_,ok := sm.EntrySuccessCounter[lastIndex]
		if(/*lastIndex == sm.NextIndex[from]*/ ok == true){
			
			if(lastIndex == sm.NextIndex[from]){
				sm.MatchIndex[from]=sm.NextIndex[from]
				sm.NextIndex[from]++
				var entriesToSend []*LogEntry
				if(int64(sm.NextIndex[from])<=sm.ThisLog.GetLastIndex()){
			
					temp1,_ := sm.ThisLog.Get(int64(sm.NextIndex[from])-1)
					persistentPrevLogEntry := temp1.(LogEntry)
					temp,_ := sm.ThisLog.Get(int64(sm.NextIndex[from]))
					persistentLogEntryToSend := temp.(LogEntry)

					arr=append(arr,MakeSendAction(from,MakeAppendEntriesReqEvent(sm.MyId,sm.CurrentTerm,sm.MyId,sm.NextIndex[from]-1,persistentPrevLogEntry.Term,append(entriesToSend,&persistentLogEntryToSend),sm.CommitIndex)))
					/*arr=append(arr,MakeSendAction(from,MakeAppendEntriesReqEvent(sm.MyId,sm.CurrentTerm,sm.MyId,sm.NextIndex[from]-1,sm.Log.	Entries[sm.NextIndex[from]-1].Term,append(entriesToSend,sm.Log.Entries[sm.NextIndex[from]]),sm.CommitIndex)))*/
				}
			}
		
			//entry_index := sm.NextIndex[from]-1
			entry_index := lastIndex
			//fmt.Println(entry_index,"*",from)
			//fmt.Println(sm.Log.Entries)
			if(int64(entry_index) <= sm.ThisLog.GetLastIndex()){
				//entry_term := sm.Log.Entries[entry_index].Term
				
				temp1,_ := sm.ThisLog.Get(int64(entry_index))
				persistentPrevLogEntry := temp1.(LogEntry)
				entry_term := persistentPrevLogEntry.Term
				map_mutex.Lock()
				//fmt.Println("here",entry_index,"+",sm.CommitIndex,"+",entry_term,"+",sm.CurrentTerm)	
				if(entry_term == sm.CurrentTerm && entry_index > sm.CommitIndex	) {	
									//fmt.Println("here",entry_index,"+",sm.CommitIndex)

					sm.EntrySuccessCounter[entry_index]++
					majority := len(sm.KnownPeers)/2 +1
					if(sm.EntrySuccessCounter[entry_index] >= majority && sm.CommitIndex < entry_index) {
						//commit this entry
						/*if sm.CommitIndex < entry_index{
							fmt.Println("",sm.CommitIndex," ",entry_index)
						}*/
						for i:=sm.CommitIndex + 1 ;i<=entry_index; i++{
							tempData,_ := sm.ThisLog.Get(int64(i))
							dataToCommit := tempData.(LogEntry)
							 
							arr=append(arr, MakeCommitAction(i,dataToCommit.Data,false))
							//fmt.Println("Commiting:",i)
						}
						sm.CommitIndex = entry_index
						//fmt.Println(sm.EntrySuccessCounter)
						//fmt.Println("Commited at index ",sm.CommitIndex," and data:",string(persistentPrevLogEntry.Data))
						//arr=append(arr, MakeCommitAction(sm.CommitIndex,persistentPrevLogEntry.Data,false))
						//fmt.Println("Leader ",sm.MyId)
						//remove this entry from map
						//delete(sm.EntrySuccessCounter,entry_index)
					}
				}
				map_mutex.Unlock()
			}
		}
	}
	
	
	//sm.ActionChannel <- arr
	return arr
}

func (sm *StateMachine) leaderVoteResp(ev *Event) ([]*Action) {

	var arr []*Action
	arr = append(arr, MakeIgnoreAction())
	//sm.ActionChannel <- arr
	return arr
}

func dummy(){
	fmt.Println()
}
