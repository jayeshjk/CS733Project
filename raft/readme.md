*Imp Note: Takes around 43 seconds to perform all the tests.

* package raft

	This package includes implementation of raft consensus protocol. Currently it uses MockCluster package for server abstraction, but modularity of code is such that new server abstraction can be added. Package provides 'RaftNode' object, which is initialized using configuration parameters defined below. The RaftNode internally contains it's own StateMachine to manage the raft protocol, and server implementation to communicate with peers in the cluster. The information about peers is provided as config parameter while initialization of RaftNode.
	To communicate with client of the raft package, currently CommitInfo channel is used. Apart from that RaftNode implements following interface, which the client can use as per requirement:
	
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

* Usage:
1) To Create the RaftNode rn = New(*NodeConfig,*mock.MockCluster) is used. Definition of NodeConfig is found in file raft_node.go .
2) Spawn the Raft functionality in separate goroutine as "go rn.RiseAndShine"
3) Use functions of Node interface to interact with this raft node.
4) See basic_test.go for initialization and usage example.s

* TODO:
Currently, comitted data comes on CommitChannel of leader node only. An initial implementation for getting this info out of follower nodes is under work.
    
