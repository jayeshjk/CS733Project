package raft

import (
	"testing"
	"os"
	//"fmt"
)

/*
 * 
 * 
 * */
var sm *StateMachine

func TestMain(m *testing.M){

	//Initialize the state machine. Get it up and running.

	//sm = MakeSM()
	//fmt.Println(sm)
	//go 	sm.RiseAndShine()
	os.Exit(m.Run())
}
func TestCommunication(t *testing.T) {

	sm = MakeSM()
	//fmt.Println(sm)
	ev := MakeAppendEvent(nil)
	//sm.EventChannel <- ev
	//a := <- sm.ActionChannel
	a := sm.RouteEvent(ev)
	a_type := a[0].ActionType
	a_success := a[0].Commit.Err
	
	if(a_type != "commit" || a_success != true || len(a)!=2 || a[1].ActionType != "commit") {
		
		t.Fatal(a_type)
	}
}
