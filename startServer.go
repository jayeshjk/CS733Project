package main

import(
	"os"
	"fmt"
	"bufio"
	"github.com/jayeshk/cs733/assignment4/raft"
	"github.com/jayeshk/cs733/assignment4/fs"
	"github.com/cs733-iitb/cluster"	
	"strconv"
)

func main(){
	args := os.Args
	//fmt.Println("here")
	file, err := os.Open(args[2])
    if err != nil {
        fmt.Println("Error opening file ++",args[1])
        return
    }
    defer file.Close()
    os.Mkdir("files",0777)
	var peers []cluster.PeerConfig
    scanner := bufio.NewScanner(file)
    m := make(map[int]int)
    for scanner.Scan() {
        id,_ := strconv.Atoi(scanner.Text())
        scanner.Scan()
        port := scanner.Text()
        scanner.Scan()
        servPort,_ := strconv.Atoi(scanner.Text())
        m[id] = servPort
        //fmt.Println(id," ",port)
        temp := cluster.PeerConfig{Id:id, Address:"localhost:"+port}
        peers = append(peers,temp)
    }
    config := cluster.Config{Peers: peers}
    thisId,_:= strconv.Atoi(args[1])
    rnode := raft.New(&raft.NodeConfig{ClusterInfo:config,Id: thisId,CurrentTermFile:"files/TermOf"+args[0],VotedForFile:"files/VotedForOf"+args[0],ElectionTimeoutControl:(thisId-1)*2000},nil,true)
    
    server := fs.InitServer(rnode.ThisConfig.Id,rnode)
    server.InitServerInfo(m)
    server.StartServerAndRaft()
}
