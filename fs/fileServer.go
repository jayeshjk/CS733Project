package fs

import(
	"github.com/jayeshk/cs733/assignment4/raft"
	"net"
	//"errors"
	"bufio"
	"encoding/json"
	"sync"
	"fmt"
	"strconv"
)

/*
 * This file contains definition and behaviour of the server in distributed file system.  
 * Server contains file system, and below it the raft layer.
 * It listens on the specified port, and handles client requests.
 * If it is a leader in the underlying raft layer, then client requests are first appended
 * to it's local log and then given to raft layer, which replicates them onto logs of other
 * servers in the cluster. A separate goroutine waits for commit info from raft layer and then
 * replies back to client. 
 * If this server is not a leader then reply with new-leader-url to client.
 * */

type ServerInfo struct{
	
	Port int
		
}

type FileServer struct{
	
	//ID
	Id int
	
	//The FileSystem
	FSys *FS
	
	//The Raft layer
	RNode *raft.RaftNode
	
	//Information about other servers in cluster
	Servers map[int]*ServerInfo
	serversLock  sync.Mutex
	
	StopAccept bool
	StopCommit bool
	
	//To differentiate between many clients
	CurrentClientId int
	ClientChannels map[int] chan *Msg
	
	wg sync.WaitGroup
	
}

func InitServer(id int,rnode *raft.RaftNode) (*FileServer){
	
	serv := FileServer{Id:id}
	serv.FSys = &FS{dir: make(map[string]*FileInfo),gversion:0}
	serv.RNode = rnode
	serv.Servers = make(map[int]*ServerInfo)
	serv.StopAccept = false
	serv.StopCommit = false
	serv.CurrentClientId = 0
	serv.ClientChannels = make(map[int] chan *Msg)
	return &serv
}

func (serv *FileServer)InitServerInfo(m map[int]int){
	//serversLock.Lock()
	for k := range m{
		
		serv.Servers[k] = &ServerInfo{Port:m[k]}
		
	}
	//serversLock.Unlock()
}

func (serv *FileServer) StartServerAndRaft(){
	
	//use waitgroup to let child goroutines run to their completion
		
	serv.wg.Add(1)
	go serv.serverMain()
	
	serv.wg.Add(1)
	go serv.commitListener()
	
	serv.wg.Add(1)
	go serv.RNode.RiseAndShine()
	
	//fmt.Println("Everything started")
	serv.wg.Wait()

}

func (serv *FileServer) serverMain(){
	
	defer serv.wg.Done()
	//serversLock.Lock()
	port := serv.Servers[serv.Id].Port
	//serversLock.Unlock()
	url := "localhost:"+strconv.Itoa(port)
	ln, err := net.Listen("tcp",url)
	
	if(err != nil) {
		
		fmt.Println("Error:",err.Error())
	
	} else{
		
		for serv.StopAccept != true {
			
			conn, err := ln.Accept() //wait for clients to connect
				
			if(err != nil){
				
				fmt.Println("Error: ",err.Error())
				
			} else {
				
				serv.CurrentClientId += 1
				tempChan := make(chan *Msg,100)
				serv.ClientChannels[serv.CurrentClientId] = tempChan
				go serv.clientListener(conn,serv.CurrentClientId,tempChan) //spawn a goroutine for client
			
			}
		}//end of for
	}//end of else	
	
	ln.Close()
}

func (serv *FileServer) clientListener(conn net.Conn,id int,commitChan chan *Msg){
	
	connReader := bufio.NewReader(conn)
	leader := serv.RNode.LeaderId()
	if(leader != serv.Id){
		//return error
		//fmt.Println("I am not a leader,Leader is: ",leader)
		_,ok := serv.Servers[leader]
		reply := ""
		if(ok){
			reply = "ERR_REDIRECT "+strconv.Itoa(serv.Servers[leader].Port)+"\r\n"
		} else {
			reply = "ERR_REDIRECT -1\r\n"
		}
		conn.Write([]byte(reply))
		delete(serv.ClientChannels,id)
		conn.Close()
		return
	}
	//fmt.Println("Connection established with leader for ",id, "my id is:",serv.Id)
	//conn.Write([]byte("OK\r\n")) // Let the client know it has reached the leader.
	count := 0

	for{
		
		leader = serv.RNode.LeaderId()
		if(leader != serv.Id){
			//return error
			//fmt.Println("I am not a leader,Leader is: ",leader)
			reply := ""
			if(leader != -1){
				reply = "ERR_REDIRECT "+strconv.Itoa(serv.Servers[leader].Port)+"\r\n"
			} else {
				reply = "ERR_REDIRECT -1\r\n"
			}
			conn.Write([]byte(reply))
			delete(serv.ClientChannels,id)	
			conn.Close()
			break
		}		
		msg, msgerr, fatalerr := GetMsg(connReader)
		if fatalerr != nil || msgerr != nil {
			serv.reply(conn, &Msg{Kind: 'M'})
			delete(serv.ClientChannels,id)
			conn.Close()
			break
		}

		if msgerr != nil {
			if (!serv.reply(conn, &Msg{Kind: 'M'})) {
				delete(serv.ClientChannels,id)
				conn.Close()
				break
			}
		}
		
		if msgerr == nil && fatalerr == nil {
			
			if msg.Kind == 'r'{
				
				response := serv.FSys.ProcessMsg(msg)
				if !serv.reply(conn, response) {
					delete(serv.ClientChannels,id)
					conn.Close()
					break
				}
							
			} else{
				msg.ClientId = id
				data,_ := json.Marshal(msg)
				//check if this server is leader
				if(serv.RNode.SM.CurrentState == "leader"){
					//Append this data to log
					serv.RNode.Lock()
					serv.RNode.Append(data)
					serv.RNode.Unlock()
					//fmt.Println("In leader above: ",id)

					response := <-commitChan
					//fmt.Println("In leader below: ",id)

					//response := serv.FSys.ProcessMsg(new_msg)
					if !serv.reply(conn, response) {
						conn.Close()
						break
					}
					count++
					if count==10{
					//fmt.Println("Response sent to ",id," count: ",count)
					}
				}else{
				//return error message
				
				}
			
			}
		}
	}
	delete(serv.ClientChannels,id)
	//fmt.Println("Connection closed for client ",id)
}

func (serv *FileServer) commitListener(){

	defer serv.wg.Done()

	//fmt.Println("From CommitListener of ",serv.Id)
	testmap := make(map[int]int)
	for{
		ci := <- serv.RNode.CommitChannel()
		if ci.Err == false {
			m1 := Msg{}
			json.Unmarshal(ci.Data,&m1)
			//fmt.Println("Recieved and processed from commitInfoChan ",string(ci.Data)," my id is ",serv.Id)
			
			response := serv.FSys.ProcessMsg(&m1)
			if(serv.RNode.SM.CurrentState == "leader"){
				_,ok := testmap[m1.ClientId]
				if !ok{
					testmap[m1.ClientId] = 1 
				} else{
					testmap[m1.ClientId]++
				}
				serv.ClientChannels[m1.ClientId] <-response
				//fmt.Println("",len(serv.ClientChannels))
				//fmt.Println("testmap:",testmap[m1.ClientId])

			}
			
		}else{
			fmt.Println("ci.Error true was received.")
		}

	}
}

func (serv *FileServer) Shutdown(){
	
	serv.StopAccept = true
	serv.StopCommit = true
	serv.RNode.Shutdown()
	defer serv.wg.Done()

}
func (serv *FileServer)reply(conn net.Conn, msg *Msg) bool {
	var err error
	var crlf = []byte{'\r', '\n'}

	write := func(data []byte) {
		if err != nil {
			return
		}
		_, err = conn.Write(data)
	}
	var resp string
	switch msg.Kind {
	case 'C': // read response
		resp = fmt.Sprintf("CONTENTS %d %d %d", msg.Version, msg.Numbytes, msg.Exptime)
	case 'O':
		resp = "OK "
		if msg.Version > 0 {
			resp += strconv.Itoa(msg.Version)
		}
	case 'F':
		resp = "ERR_FILE_NOT_FOUND"
	case 'V':
		resp = "ERR_VERSION " + strconv.Itoa(msg.Version)
	case 'M':
		resp = "ERR_CMD_ERR"
	case 'I':
		resp = "ERR_INTERNAL"
	default:
		fmt.Printf("Unknown response kind '%c'", msg.Kind)
		return false
	}
	resp += "\r\n"
	write([]byte(resp))
	if msg.Kind == 'C' {
		write(msg.Contents)
		write(crlf)
	}
	return err == nil
}
