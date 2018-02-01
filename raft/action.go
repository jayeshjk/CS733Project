package raft

import(

)

/*
 * Implemetation of Actions by State Machine.
 * Similar to implementation of Event
 * */

type ApplyData struct {
	//Apply the command in Log at Index  
	Index int
}
 
type SendData struct {
	
	PeerId int
	EventToSend *Event

}

type CommitData struct {

	Index int
	Data []byte
	Err bool
}

type AlarmData struct {

	Timeout int

}

type LogStoreData struct {

	Index int
	//Data []byte
	Entry *LogEntry
}

type LogTruncateData struct {
	
	FromIndex int

}

type WriteToFileData struct {
	Name string
	Data int
}

type Action struct{
	
	ActionType string
	Send *SendData
	Commit *CommitData
	Alarm *AlarmData
	LogStore *LogStoreData
	LogTruncate *LogTruncateData
	Write *WriteToFileData
	Apply *ApplyData
}
