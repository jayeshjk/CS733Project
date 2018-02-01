package raft

import (
	"sync"
	"github.com/cs733-iitb/log"
)

/*
 * This file contains definition of RAFT State Machine.
 * Following struct definition captures the state of the
 * machine, both persistent and volatile, and also various 
 * operations on the state machine are implemented below.
 * */
 /*
 var FIX_PART int = 10
 var CANDIDATE_TIMEOUT = 10
 var ELECTION_TIMEOUT int = 5
 var ELECTION_OVER int = 15
 var HEARTBEAT_TIMEOUT int = 5
*/

 var FIX_PART int = 500
 var CANDIDATE_TIMEOUT = 1000
 var ELECTION_TIMEOUT int = 1000
 var ELECTION_OVER int = 15
 var HEARTBEAT_TIMEOUT int = 250
 var map_mutex = &sync.Mutex{}
 
 type LogEntry struct {
	 
	 Term int
	 Data []byte
 
 }
  
type StateMachine struct {
	
	sync.Mutex
	
	/*
	 * Persistent data. 
	 * TODO: Add persistent storage support 
	 * */
	
	CurrentState string /* leader / follower / candidate (all in smalls)*/
	CurrentTerm int   /* Latest Term this machine has seen*/
	VotedFor int        /* CandidateId of server this macine voted in CurrentTerm*/

	//persistent log
	ThisLog *log.Log	
	
	/*
	 * Volatile data
	 * */
	MyId int
	CommitIndex int     /* Index of highest committed log entry.*/
	KnownLeaderId int   /* Index of the latest known leader */
	KnownPeers []int
	
	/*
	 * Data specific to candidate state
	 * */
	VoteCounter int    /* Number of votes received by this candidate.*/
	NegativeVoteCounter int 
	VoteReceived map[int]int
	/*
	 * Data specific to leader state
	 * */
	NextIndex map[int] int   /* To keep track of entries sent to followers */
	MatchIndex map[int] int /* Index of highest entry replicated on server*/
	EntrySuccessCounter map[int] int /* To make commit related decisions*/
	LogRepairing map[int] bool /* To keep track whether log is being repaired or not */
		 
}
