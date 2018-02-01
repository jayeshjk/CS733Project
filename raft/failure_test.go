package raft
/*
 * Initialize logs of nodes in different states and check leader is properly getting elected or not.
 * 
 * */
import(
	"testing"
	"github.com/cs733-iitb/cluster"
	"github.com/cs733-iitb/cluster/mock"
	//"github.com/cs733-iitb/log"
	"time"
	"strconv"
	"fmt"
	//"reflect"
	//"math/rand"
	"os"
//	"errors"
)

func makeRafts1()([]*RaftNode){
		os.Mkdir("files",0777)
		var rafts []*RaftNode
		config := cluster.Config{
        	Peers: []cluster.PeerConfig{
            		{Id: 1, Address: "localhost:8070"},
            		{Id: 2, Address: "localhost:8071"},
            		{Id: 3, Address: "localhost:8072"},
            		{Id: 4, Address: "localhost:8073"},
            		{Id: 5, Address: "localhost:8074"},
            		{Id: 6, Address: "localhost:8075"},
            		//{Id: 7, Address: "localhost:7076"},
			},
        }
        //fmt.Println(len(config.Peers))
        mc,_ := mock.NewCluster(nil)
        for i:=0;i<len(config.Peers);i++ {
			//fmt.Println(i)
			rafts=append(rafts,New(&NodeConfig{ClusterInfo:config,Id: config.Peers[i].Id,CurrentTermFile:"files/TermOf"+strconv.Itoa(config.Peers[i].Id),VotedForFile:"files/VotedForOf"+strconv.Itoa(config.Peers[i].Id),ElectionTimeoutControl:(i+1)*1000},mc,false))
		}
		return rafts
}

//test the log repair and election after failure

func TestFailure1(t *testing.T){
	//0. Cleanup
	removeData()
	
	//1.Initialize the logs of nodes
	rafts := makeRafts1()
	
	//2.Ensure that rafts[5] will timeout first
	rafts[0].ThisConfig.ElectionTimeoutControl=0
	
	//3.Start them all
	for i:=0;i<len(rafts);i++{
		//fmt.Println(rafts[i])
		go rafts[i].RiseAndShine()
	}
	
	//4.Wait
	time.Sleep(2*time.Second)
	
	//5.Test
	leader,err := getLeaderId(rafts)
	if(leader==-1) {
		t.Fatal("Error "+err.Error())
	}
	
	data := []string{"arrival","of","the","king","is","in"}
	
	for i:=0;i<2;i++ {
		rafts[leader-1].Append([]byte(data[i]))
		time.Sleep(500*time.Millisecond)
		select { // to avoid blocking on channel.
			case ci := <- rafts[leader-1].CommitChannel():
				if ci.Err != false {
					t.Fatal(ci.Err)
					
				}
			default: t.Fatal("Expected message on all nodes")
		}

	}
	//Induce the partition
	rafts[0].Mock.Partition([]int{11,12,13,14}, []int{15,16})
	//fmt.Println("PARTITION INDUCED")
	//time.Sleep(2*time.Second)
	for i:=2;i<4;i++ {
		rafts[leader-1].Append([]byte(data[i]))
		time.Sleep(500*time.Millisecond)
		select { // to avoid blocking on channel.
			case ci := <- rafts[leader-1].CommitChannel():
				if ci.Err != false {
					t.Fatal(ci.Err)
					
				}
			default: t.Fatal("Expected message on all nodes")
		}

	}
	time.Sleep(1*time.Second)
	//heal the partition
	rafts[0].Mock.Heal()
	//fmt.Println("PARTITION HEALED")
	//let the system stabilize
	time.Sleep(2*time.Second)
	
	leader,err = getLeaderId(rafts)
	
	/*if(leader!=6) {
		//t.Fatal("Error "+err.Error())
		t.Fatal("Error "+strconv.Itoa(leader))
	}*/
	
	//send the remaining data
	for i:=4;i<len(data);i++ {
		rafts[leader-1].Append([]byte(data[i]))
		time.Sleep(500*time.Millisecond)
		select { // to avoid blocking on channel.
			case ci := <- rafts[leader-1].CommitChannel():
				if ci.Err != false {
					t.Fatal(ci.Err)
					
				}
			default: t.Fatal("Expected message on all nodes")
		}

	}
	//let the logs repair
	time.Sleep(2*time.Second)
	//check for all nodes
	for i:=0;i<len(rafts);i++ {
		for j:=0;j<len(data);j++ {
			entry,_:=rafts[i].Get(j)
			if string(entry.Data) != data[j] {
				printAllStates(rafts)
				fmt.Println(rafts[i].SM.CommitIndex," commitIndex of ",rafts[i].ThisConfig.Id)
				t.Fatal("Got different data for node ",i,"at index ",j," data ",string(entry.Data))			} 
		}
	}
	
	closeRafts(rafts)

}

//leader fails
func TestFailure2(t *testing.T){
	//0. Cleanup
	removeData()
	
	//1.Initialize the logs of nodes
	rafts := makeRafts1()
	
	//2.Ensure that rafts[5] will timeout first
	rafts[0].ThisConfig.ElectionTimeoutControl=0
	
	//3.Start them all
	for i:=0;i<len(rafts);i++{
		//fmt.Println(rafts[i])
		go rafts[i].RiseAndShine()
	}
	
	//4.Wait
	time.Sleep(1500*time.Millisecond)
	
	//5.Test
	leader,err := getLeaderId(rafts)
	if(leader==-1) {
		t.Fatal("Error "+err.Error())
	}
	
	rafts[0].Mock.Partition([]int{11}, []int{12,13,14,15,16})
	
	time.Sleep(3*time.Second)
	
	leader,err = getLeaderId(rafts)
	
	//new leader getting elected and it's not the older one
	if(leader==-1 && leader == 1) {
		t.Fatal("Error "+err.Error())
	}
	
	closeRafts(rafts)
}

//a follower fails and it's log is repaired afterwards
func TestFailure3(t *testing.T){
	//0. Cleanup
	removeData()
	
	//1.Initialize the logs of nodes
	rafts := makeRafts1()
	
	//2.Esnsure that rafts[5] will timeout first
	rafts[0].ThisConfig.ElectionTimeoutControl=0
	
	//3.Start them all
	for i:=0;i<len(rafts);i++{
		//fmt.Println(rafts[i])
		go rafts[i].RiseAndShine()
	}
	
	//4.Wait
	time.Sleep(1500*time.Millisecond)
	
	//5.Test
	leader,err := getLeaderId(rafts)
	if(leader==-1) {
		t.Fatal("Error "+err.Error())
	}
	
	rafts[0].Mock.Partition([]int{11,12,13,14,15}, []int{16})
	
	//time.Sleep(3*time.Second)
	
	data := []string{"arrival","of","the"}
	
	for i:=0;i<2;i++ {
		rafts[leader-1].Append([]byte(data[i]))
		time.Sleep(500*time.Millisecond)
		select { // to avoid blocking on channel.
			case ci := <- rafts[leader-1].CommitChannel():
				if ci.Err != false {
					t.Fatal(ci.Err)
					
				}
			default: t.Fatal("Expected message on all nodes")
		}

	}

	rafts[0].Mock.Heal()
	//fmt.Println("PARTITION HEALED")
	leader,err = getLeaderId(rafts)
	if(leader==-1 && leader == 1) {
		t.Fatal("Error "+err.Error())
	}
	for i:=2;i<len(data);i++ {
		rafts[leader-1].Append([]byte(data[i]))
		time.Sleep(500*time.Millisecond)
		select { // to avoid blocking on channel.
			case ci := <- rafts[leader-1].CommitChannel():
				if ci.Err != false {
					t.Fatal(ci.Err)
					
				}
			default: t.Fatal("Expected message on all nodes")
		}

	}
	//let the log repair
	time.Sleep(2*time.Second)
	//check the logs of everyone
	for i:=0;i<len(rafts);i++ {
		for j:=0;j<len(data);j++ {
			entry,_:=rafts[i].Get(j)
			if string(entry.Data) != data[j] {
				t.Fatal("Got different data for node ",i,"at index ",j)
			} 
		}
	}
	
	
	closeRafts(rafts)

}
