package fs

import(
	. "github.com/jayeshk/cs733/assignment4/raft"
	"testing"
	"github.com/cs733-iitb/cluster"
	"github.com/cs733-iitb/cluster/mock"
	"time"
	"strconv"
	"fmt"
	//"reflect"
	//"math/rand"
	"os"
	"errors"
	"net"
	"bufio"
	"strings"
	"sync"
)

var wg_write sync.WaitGroup
var rafts []*RaftNode
var servers []*FileServer

func getLeaderId()(int,error){
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

func printAllStates(){
	
	for i:=0;i<len(rafts);i++ {
		fmt.Println(rafts[i].SM.CurrentState," ",rafts[i].ThisConfig.Id)	
	}
}

func closeServers(){
	if(servers[0].RNode.Mock != nil){
		servers[0].RNode.Mock.Close()
	}
	for i:=0;i<len(servers);i++{
		servers[i].Shutdown()
	}
}

//For cleanup
func removeData(){
	os.RemoveAll("Logs")
	os.RemoveAll("files")
	//os.Mkdir("files",0777)
}

func makeRafts(){
	
		os.Mkdir("files",0777)
		//var rafts []*RaftNode
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
			rafts=append(rafts,New(&NodeConfig{ClusterInfo:config,Id: config.Peers[i].Id,CurrentTermFile:"files/TermOf"+strconv.Itoa(config.Peers[i].Id),VotedForFile:"files/VotedForOf"+strconv.Itoa(config.Peers[i].Id),ElectionTimeoutControl:(i+1)*2000},mc,false))
		}
		//return rafts
}

func makeServers(){

	//var servers []*FileServer
	//create map of ports
	m := make(map[int]int)
	initialPort := 5000
	for i:=0;i<len(rafts);i++{
		
		m[rafts[i].ThisConfig.Id] = initialPort
		initialPort += 1
	}
	
	for i:=0;i<len(rafts);i++ {
		
		servers = append(servers,InitServer(rafts[i].ThisConfig.Id,rafts[i]))
		servers[i].InitServerInfo(m)
	}
	
	//return servers
}

func TestMain(m *testing.M){
	//go serverMain()
	//time.Sleep(1*time.Second)
	//0. Cleanup
	removeData()
	
	//1.Initialize the logs of nodes
	makeRafts()
	
	//2.Ensure that rafts[5] will timeout first
	rafts[4].ThisConfig.ElectionTimeoutControl=0
	
	makeServers()
	
	//start all servers
	for i:=0;i<len(servers);i++ {
		go servers[i].StartServerAndRaft()
	}
	
	//let the system stabilize
	time.Sleep(2*time.Second)
	//m.Run()
	os.Exit(m.Run())
}

func TestClusterFormation(t *testing.T){
	
	//get the id of leader from server
	//start a client
	conn, _ := net.Dial("tcp","127.0.0.1:5000")
	connReader := bufio.NewReader(conn)
	
	response,_ := connReader.ReadString('\n')
	
	if(!strings.Contains(response,"ERR_REDIRECT")){
		t.Errorf("Error:",response)
	}
	
	//closeServers(servers)
	//time.Sleep(1*time.Second)
	
	conn.Close()

	url := "127.0.0.1:"+strings.Fields(response)[1]
	//leaderPort,_ := strconv.Atoi(strings.Fields(response)[1])
	conn, _ = net.Dial("tcp",url)
	connReader = bufio.NewReader(conn)

	conn.Write([]byte("write myfile 10 100\r\nabc\x00\xFF\x011234\r\n"))
	response,_ = connReader.ReadString('\n')
	//fmt.Println(response)
	if (!strings.Contains(response,"OK")){
		fmt.Println(response)
		t.Errorf("Error")
	}
	
	time.Sleep(1*time.Second)
	//check for maps of all servers
	/*for i:=0;i<len(servers);i++ {
		
		_,ok := servers[i].FSys.dir["myfile"]
		if ok == false {
			fmt.Print(servers[i].FSys.dir)
			t.Errorf("File not found at server ",i)
		}
		//TODO:check the contents also 
	}*/
	
	conn.Write([]byte("read myfile\r\n"))
	response,_ = connReader.ReadString('\n')
	contents,_ := connReader.ReadString('\n') //Beware it might get hanged here
	//fmt.Println(string(contents))
	if (!strings.Contains(response,"CONTENTS")){
		fmt.Println(string(contents))
		t.Errorf("Error")
	}
}

/*
func oneWriteClient(id int,url string){
	//<-ch
	fmt.Println("Dialing to port 8080")
	conn, _ := net.Dial("tcp",url)
	connReader := bufio.NewReader(conn)
	//data :=make([]string,10)
	defer wg_write.Done()
	data:=[]string{"a","b","c","d","e","f","g","h","i","j"}
	i:=0
	for ;i<len(data);i++ {
		conn.Write([]byte("write multi-file1 1 10000\r\n"+data[i]+"\r\n"))
		response,_ := connReader.ReadString('\n')
		//fmt.Println("Client: ",id," iteration: ",i," response: ",response)
		if (!strings.Contains(response,"OK")){
			fmt.Println(response)
			//t.Errorf("Error")
		}
	}
	conn.Close()
}

func TestConcWrites(t *testing.T){
	
	conn, _ := net.Dial("tcp","127.0.0.1:5000")
	connReader := bufio.NewReader(conn)
	
	response,_ := connReader.ReadString('\n')
	
	if(!strings.Contains(response,"ERR_REDIRECT")){
		t.Errorf("Error:",response)
	}
	
	//closeServers(servers)
	//time.Sleep(1*time.Second)
	
	conn.Close()

	url := "127.0.0.1:"+strings.Fields(response)[1]

	conn, _ = net.Dial("tcp",url)
	connReader = bufio.NewReader(conn)

	const no int = 10	
	i:=0
	for ;i<no;i++ {
		//children_channels[i] <- 0
		wg_write.Add(1)
		go oneWriteClient(i,url)
	}
	
	wg_write.Wait()
	time.Sleep(1*time.Second)
	//All children died
	conn.Write([]byte("read multi-file1\r\n "))
	message,_ := connReader.ReadString('\n')
	message,_ = connReader.ReadString('\n')
	
	if(!strings.Contains(message,"j")){
		t.Errorf("Error "+message)
	}	
	conn.Close()	
}
*/
