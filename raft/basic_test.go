package raft
/*
 * Initialize logs of nodes in different states and check leader is properly getting elected or not.
 * 
 * */
import(
	"testing"
	"github.com/cs733-iitb/cluster"
	"github.com/cs733-iitb/cluster/mock"
//	"github.com/cs733-iitb/log"
	"time"
	"strconv"
	"fmt"
	//"reflect"
	//"math/rand"
	"os"
	"errors"
)

func getLeaderId(rafts []*RaftNode)(int,error){
	leader := -1
	for i:=0;i<len(rafts);i++{
		if(leader == -1){
			leader = rafts[i].LeaderId()
		} else if(leader != rafts[i].LeaderId()) {
			return -1,errors.New("Different leaders")
		}
	}
	return leader,nil
}

func printAllStates(rafts []*RaftNode){
	
	for i:=0;i<len(rafts);i++ {
		fmt.Println(rafts[i].SM.CurrentState," ",rafts[i].ThisConfig.Id)	
	}
}

func closeRafts(rafts []*RaftNode){
	if(rafts[0].Mock != nil){
		rafts[0].Mock.Close()
	}
	for i:=0;i<len(rafts);i++{
		rafts[i].Shutdown()
	}
}

//For cleanup
func removeData(){
	os.RemoveAll("Logs")
	os.RemoveAll("files")
	//os.Mkdir("files",0777)
}

func makeRafts()([]*RaftNode){
	
		os.Mkdir("files",0777)
		var rafts []*RaftNode
		config := cluster.Config{
        	Peers: []cluster.PeerConfig{
            		{Id: 1, Address: "localhost:7070"},
            		{Id: 2, Address: "localhost:7071"},
            		{Id: 3, Address: "localhost:7072"},
            		{Id: 4, Address: "localhost:7073"},
            		{Id: 5, Address: "localhost:7074"},
            		//{Id: 6, Address: "localhost:7075"},
			},
        }
        //fmt.Println(len(config.Peers))
        mc,_ := mock.NewCluster(nil)
        for i:=0;i<len(config.Peers);i++ {
			//fmt.Println(i)
			rafts=append(rafts,New(&NodeConfig{ClusterInfo:config,Id: config.Peers[i].Id,CurrentTermFile:"files/TermOf"+strconv.Itoa(config.Peers[i].Id),VotedForFile:"files/VotedForOf"+strconv.Itoa(config.Peers[i].Id),ElectionTimeoutControl:(i+1)*1000+10000},mc,false))
		}
		return rafts
}

/*
func TestElection1(t *testing.T){
	
	//When Majority votes are possible
	//0. Cleanup
	removeData()
	
	//1.Initialize the logs of nodes
	rafts := makeRafts()
	
	//2.Ensure that rafts[5] will timeout first
	rafts[4].ThisConfig.ElectionTimeoutControl=0
	
	//3.Start them all
	for i:=0;i<len(rafts);i++{
		fmt.Println(rafts[i])
		go rafts[i].RiseAndShine()
	}
	
	//4.Wait
	time.Sleep(4*time.Second)
	
	//5.Test
	leader,_ := getLeaderId(rafts)
	if(leader==-1 || leader != rafts[4].ThisConfig.Id) {
		t.Fatal("Error ",leader)
	}
	
	//6.Close rafts
	closeRafts(rafts)		
}
*/
func TestAppend(t *testing.T){
	//0. Cleanup
	removeData()
	
	//1.Initialize the logs of nodes
	rafts := makeRafts()
	
	//2.Ensure that rafts[5] will timeout first
	rafts[4].ThisConfig.ElectionTimeoutControl=0
	
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
	
	//data := []string{"arrival","of"}
	//data := []string{"arrival","of","the","king","is","in","the","prophecy"}
	totalDataToSend := 10
	for i:=0;i<totalDataToSend/*len(data)*/;i++ {
		//time.Sleep(1*time.Second) //time to stabilize the system
		rafts[leader-1].Append([]byte(strconv.Itoa(i)))
		//time.Sleep(700*time.Millisecond)
		if(i==0){
			ci := <- rafts[leader-1].CommitChannel()
				if ci.Err != false {
					printAllStates(rafts)
					t.Fatal(ci.Err," ",i)	
				}
		} else {
			for j:=0;j<len(rafts);j++ {
				ci := <- rafts[j].CommitChannel()
				if ci.Err != false {
					printAllStates(rafts)
					t.Fatal(ci.Err," ",i)
					
				}
				//fmt.Println("commited:",j)
			}
		}
		//fmt.Println("data",i)
		//printAllStates(rafts)
		/*select { // to avoid blocking on channel.
			case ci := <- rafts[leader-1].CommitChannel():
				if ci.Err != false {
					printAllStates(rafts)
					t.Fatal(ci.Err," ",i)
					
				}
			default:
				printAllStates(rafts)
				t.Fatal("Expected message on all nodes",i)
		}*/

	}
	time.Sleep(2*time.Second)
	//check for all nodes
	for i:=0;i<len(rafts);i++ {
		for j:=0;j<totalDataToSend/*len(data)*/;j++ {
			entry,_:=rafts[i].Get(j)
			if string(entry.Data) != strconv.Itoa(j)/*data[j]*/ {
				printAllStates(rafts)
				fmt.Println(rafts[i].SM.CommitIndex," commitIndex of ",i)
				t.Fatal("Got different data for node ",i,"at index ",j," data ",string(entry.Data))
			} 
		}
	}
	//fmt.Println("==================================")
	
	closeRafts(rafts)

}
