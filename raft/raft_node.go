package raft

/*
 * This file contains basic struct definitions
 * */

import(
	"github.com/cs733-iitb/cluster"
	"github.com/cs733-iitb/cluster/mock"
	"github.com/cs733-iitb/log"
	"strconv"
	"time"
	"fmt"
	"encoding/gob"
	"io/ioutil"
	"math/rand"
	"sync"
)

type Message struct {
	
	From int
	Ev *Event
	
}

type NetConfig struct {

	Id int
	Host string
	Port int

}

type NodeConfig struct{
	
	ClusterInfo cluster.Config // Information about all servers, including this.
	Id int              // this node's id. One of the cluster's entries should match.
	ElectionTimeoutControl int
	HeartbeatTimeoutControl int
	CurrentTermFile string
	VotedForFile string
}

type CommitInfo struct {
	
	Data []byte
	Index int 
	//Err error
	
}

type RaftNode struct{
	sync.Mutex
	
	/* Configurations */
	SM *StateMachine
	ThisNode cluster.Server
	ThisConfig *NodeConfig
	ThisLog *log.Log
	Mock *mock.MockCluster
	
	/* Misc */
	CurrentMsgId int64 //Monotonically increasing Msg Id 
	
	/* Channels*/
	/* RaftNode listens on this channel. Various child goroutines receive
	 * events such as append, messages from peers, timeouts etc. and  
	 * make appropriate Event object and put on this channel.
	 * The goroutine listening on this thread will take this object out 
	 * of the channel and feed it to StateMachine's input and
	 * will wait for array of action objects, to process it later.
	 * */
	IncomingEvents chan *Event     
	CommitInfoChan chan *CommitData //output from node to client
	//struct CommitData is defined in action.go
	
	/*
	 * Timers
	 * */
	 ThisTimer *time.Timer
}

/*func (r *RaftNode) PerformRecovery(){
	for i:=0;i<=r.ThisLog.GetLastIndex();i++{
		tempData,_ := r.ThisLog.Get(int64(i))
		dataToCommit := tempData.(LogEntry)
							 
		arr = MakeCommitAction(i,dataToCommit.Data,false))

		r.CommitInfoChan <- arr.Commit
	}
}*/

func New(config *NodeConfig,mc *mock.MockCluster,useCluster bool) *RaftNode {
	//Initilize & return RaftNode with configuration in config

	sm := MakeSM()
	
	sm.Lock()
	defer sm.Unlock()

	sm.MyId = config.Id
	sm.CommitIndex=-1
	//sm.LastLogIndex=-1
	
	r := RaftNode{SM:sm, ThisConfig:config, IncomingEvents:make(chan *Event), CommitInfoChan:make(chan *CommitData,1000)}
	
	if(useCluster == false){
		r.Mock = mc
		srv,err := r.Mock.AddServer(r.ThisConfig.Id)
		//srv, err := cluster.New(r.ThisConfig.Id, config.ClusterInfo)
		if(err == nil){
			r.ThisNode = srv
		} else {
			fmt.Println("Error while initializing server.")
		}
	} else {
		srv, err := cluster.New(r.ThisConfig.Id, config.ClusterInfo)
	
		if(err == nil){
			r.ThisNode = srv
		} else {
			fmt.Println("Error while initializing server.")
		}
	}

	tempLogName := "Logs/LogOf"+strconv.Itoa(r.ThisConfig.Id)
	tempLog, err := log.Open(tempLogName)
	if(err != nil){
		fmt.Println("Error while opening the Log: "+ tempLogName)
	}else {
		r.ThisLog = tempLog
		r.ThisLog.RegisterSampleEntry(LogEntry{})
	}
	
	//Initialize sm with info in NetConfig and Log
	for i:=0;i<len(config.ClusterInfo.Peers);i++ {
		
		if(config.ClusterInfo.Peers[i].Id != config.Id){
			r.SM.KnownPeers = append(r.SM.KnownPeers,config.ClusterInfo.Peers[i].Id)
		}
	}
	//Initialize SM with persistent data
	
	r.SM.ThisLog = r.ThisLog
	/*last := r.ThisLog.GetLastIndex()
	for i:=int64(0);i<=last;i++{
		data,_ := r.ThisLog.Get(int64(i))
		tempEntry := data.(LogEntry)
		r.SM.Log.Entries[int(i)] = &tempEntry //watch out for error here
	}*/
	
	
	
	b, err := ioutil.ReadFile(r.ThisConfig.CurrentTermFile)
    if err != nil {
        ioutil.WriteFile(r.ThisConfig.CurrentTermFile, []byte("0"), 0644)
        r.SM.CurrentTerm=0
    }else{
		r.SM.CurrentTerm,_=strconv.Atoi(string(b))
	}
	
	b, err = ioutil.ReadFile(r.ThisConfig.VotedForFile)
    if err != nil {
        ioutil.WriteFile(r.ThisConfig.VotedForFile, []byte("-1"), 0644)
        r.SM.VotedFor=-1
    }else{
		r.SM.VotedFor,_=strconv.Atoi(string(b))
	}
	//fmt.Println("IN the new function")

	return &r
}

type Node interface {
	
	// Client's message to Raft node
	Append([]byte)

	// A channel for client to listen on. What goes into Append must 
	// come out of here at some point.
	CommitChannel() <- chan CommitInfo

	// Last known committed index in the log. This could be -1 until 
	// the system stabilizes.
	CommittedIndex() int

	// Returns the data at a log index, or an error.
	Get(index int) ([]byte,error)

	// Node's id
	Id() int

	// Id of leader. -1 if unknown
	LeaderId() int

	// Signal to shut down all goroutines, stop sockets, flush log and 
	// close it, cancel timers.
	Shutdown()

}

func (r *RaftNode) Append(data []byte){
	//create append event and put it on event channel
	e := MakeAppendEvent(data)
	r.IncomingEvents <- e
	//fmt.Println("Put into IncomingEvents ",string(data))
}

func (r *RaftNode) CommitChannel() <- chan *CommitData{
	
	return r.CommitInfoChan
	
}

func (r *RaftNode) CommitedIndex() int{
	//Have to handle the concurrent access
	r.SM.Lock()
	defer r.SM.Unlock()
	
	return r.SM.CommitIndex
}

func (r *RaftNode) Get(index int) (LogEntry,error){
	
	data,err := r.ThisLog.Get(int64(index))
	if(err != nil){
		//fmt.Println("r.Get",r.ThisConfig.Id)
		//fmt.Println("r.Get Error",err.Error())
		return LogEntry{},err //watch out for error here
	}else{
		//fmt.Println("r.Get",r.ThisConfig.Id)
		return data.(LogEntry),nil
	}
}

func (r *RaftNode) Id() int{
	
	return r.ThisConfig.Id
	
}

func (r *RaftNode) LeaderId() int{
	//Need to get it from StateMachine. Concurrent access should be managed
	r.SM.Lock()
	defer r.SM.Unlock()
	
	return r.SM.KnownLeaderId
}

func (r *RaftNode) Shutdown(){
	
	r.ThisTimer.Stop()
	r.IncomingEvents <- MakeStopEvent()
	//r.ThisNode.Close()
	
}


func (r *RaftNode) ListenForMessage(){
	
	for{
		envelope := <- r.ThisNode.Inbox()
		m1 := envelope.Msg.(*Message)
		ev := m1.Ev
		//fmt.Printf("Message Received %d %s\n",r.ThisConfig.Id,ev.EventType)
		//fmt.Println(ev)
		r.IncomingEvents <- ev
	}
}

func (r *RaftNode) RiseAndShine() {
	//Spawn child goroutines
	// 1. To listen on IncomingEvent channel and so on
	// 2. To listen for messages from peers
	
	gob.Register(&Message{})
	//fmt.Println("Initiating ",r.ThisConfig.Id," with timer",ELECTION_TIMEOUT+r.ThisConfig.ElectionTimeoutControl+rand.Intn(5))
	r.ThisTimer = time.AfterFunc(time.Duration(ELECTION_TIMEOUT+r.ThisConfig.ElectionTimeoutControl+rand.Intn(5))*time.Millisecond,r.handleTimeout)
	//go r.StartTimer(ELECTION_TIMEOUT+r.ThisConfig.ElectionTimeoutControl+rand.Intn(5))
	go r.ListenForMessage()
	//TODO timers
	
	for{
		//Take the event out from Incoming event channel
		ev := <- r.IncomingEvents
		
		//check for the stop event
		if(ev.EventType == "stop") {
			break
		}
		arr := r.SM.RouteEvent(ev)
		//Process actions
		for i:=0;i<len(arr);i++ {
			if(arr[i].ActionType=="send"){
				
				//send to given server.
				//Extract info from action object
				env := &cluster.Envelope{Pid:arr[i].Send.PeerId,MsgId:r.CurrentMsgId,Msg: &Message{From:r.ThisConfig.Id,Ev:arr[i].Send.EventToSend}}
				r.CurrentMsgId++
				//send the message
				//fmt.Println("before")
				r.ThisNode.Outbox() <- env
				//fmt.Println("after")
				/*if(r.SM.CurrentState == "candidate" && arr[i].Send.EventToSend.VoteReq != nil){
					fmt.Println("IN candidate ",arr[i].Send.EventToSend.From,"  sent to", env.Pid)
				}*/
				//fmt.Println("Message successfully posted for ",env.Pid,"by ",r.ThisConfig.Id)
				
			}else if(arr[i].ActionType=="commit"){
				
				//return CommitData to client
				r.CommitInfoChan <- arr[i].Commit
				//fmt.Println("Putting into commitInfoChan ",string(arr[i].Commit.Data),"Id: ",r.ThisConfig.Id," Current State: ",r.SM.CurrentState)
				 
				
			}else if(arr[i].ActionType=="alarm"){
				//Extract timeout value from action object and reset the timer.
				//fmt.Println("Action: %s Node: %d",arr[i].ActionType,r.ThisConfig.Id,r.SM.CurrentState)
				timeout := arr[i].Alarm.Timeout
				//fmt.Println(r.ThisConfig.Id)
				if(r.SM.CurrentState == "follower"){
					timeout= timeout + r.ThisConfig.ElectionTimeoutControl
				}
				//fmt.Println(timeout)
				//fmt.Println(r.ThisConfig.Id," ",r.ThisTimer.Stop())
				r.ThisTimer.Stop()
				r.ThisTimer = time.AfterFunc(time.Duration(timeout)*time.Millisecond,r.handleTimeout)
				//r.ThisTimer.Reset(time.Duration(timeout) * time.Second)
				//go r.StartTimer(timeout)
			}else if(arr[i].ActionType=="logstore"){
				//Add this to log of this node
				//fmt.Println("Action: %s Node: %d",arr[i].ActionType,r.ThisConfig.Id,r.SM.CurrentState)
				r.ThisLog.Append(*(arr[i].LogStore.Entry))
				
			} else if(arr[i].ActionType=="logtruncate"){
				//Truncate the log
				//fmt.Println("Action: %s Node: %d",arr[i].ActionType,r.ThisConfig.Id,r.SM.CurrentState)
				r.ThisLog.TruncateToEnd(int64(arr[i].LogTruncate.FromIndex))
			} else if(arr[i].ActionType=="writetofile"){
				//Truncate the log
				var tempFileName string
				if(arr[i].Write.Name == "term"){
					tempFileName = r.ThisConfig.CurrentTermFile
				} else {
					tempFileName = r.ThisConfig.VotedForFile
				}
				
				ioutil.WriteFile(tempFileName, []byte(strconv.Itoa(arr[i].Write.Data)), 0644)
			} else if(arr[i].ActionType=="apply"){
				// TODO: Here comes the code to let the client know about applying the commands.
				// For this purpose a dedicated channel or CommitInfo channel may be used.
				// This would be implemented in next iteration of assignment where a client 
				// will be fused together with raft layer.
				
			} 
		}
	}
}

func (r *RaftNode)StartTimer(timeout int){
	
	r.ThisTimer = time.AfterFunc(time.Duration(timeout)*time.Second,r.handleTimeout)
	
}
func (r *RaftNode)handleTimeout(){
	
	ev := MakeTimeoutEvent()
	//fmt.Println("Timeout happened for ",r.ThisConfig.Id)
	r.IncomingEvents <- ev
	//fmt.Println("Timeout event put in channel for ",r.ThisConfig.Id)
}
