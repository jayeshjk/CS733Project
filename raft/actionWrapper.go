package raft

import (

)

/*
 * 
 * */
 //ActionType: send, commit, alarm, logstore, ignore
func MakeSendAction(peerId int, eventToSend *Event) *Action{
	 a := Action{ActionType:"send", Send: &SendData{PeerId:peerId,EventToSend:eventToSend}}
	 return &a
} 
 
 func MakeCommitAction(index int,data []byte,err bool) *Action {
	 a := Action{ActionType:"commit", Commit: &CommitData{Index:index,Data:data,Err:err}}
	 return &a
}
 
func MakeAlarmAction(timeout int) *Action {
	 a := Action{ActionType:"alarm", Alarm: &AlarmData{Timeout:timeout}}
	 return &a
 }
 
func MakeLogStoreAction(index int, entry *LogEntry) *Action {
	 
	 a := Action{ActionType:"logstore", LogStore: &LogStoreData{Index:index,Entry:entry}}
	 return &a
}

func MakeLogTruncateAction(index int) *Action {
	 a := Action{ActionType:"logtruncate", LogTruncate: &LogTruncateData{FromIndex:index}}
	 return &a
 }
 
 func MakeWriteToFileAction(name string,data int) *Action {
	 a := Action{ActionType:"writetofile", Write: &WriteToFileData{Name:name,Data:data}}
	 return &a
 }
 
 func MakeApplyAction(index int) *Action {
	 a := Action{ActionType:"apply", Apply: &ApplyData{Index:index}}
	 return &a
 }
 
func MakeIgnoreAction() *Action{
	
	a := Action{ActionType:"ignore"}
	return &a

}
