package main
import(
	"os"
	"testing"
//	"errors"
	"fmt"
	"strconv"
	"net"
	"time"
	"bufio"
	"strings"
)
var servers []*os.Process
var total = 6
var ports =[]int{5000,5001,5002,5003,5004,5005}
var path_to_config = "$GOPATH/src/github.com/jayeshk/cs733/assignment4/config"
var path_to_binary = "$GOPATH/bin/assignment4"

//var leaderPort int
//For cleanup
func removeData(){
	os.RemoveAll("Logs")
	os.RemoveAll("files")
	//os.Mkdir("files",0777)
}

func findLeaderPort(ports []int) string {

	for i:=0;i<len(ports);i++{
		tempId := strconv.Itoa(ports[i])
		//fmt.Println(tempId)

		url := "127.0.0.1:"+tempId
		conn, err := net.Dial("tcp",url)
		if(err== nil){
			connReader := bufio.NewReader(conn)
			conn.Write([]byte("read dummy\r\n"))
			response,_ := connReader.ReadString('\n')
			//fmt.Println("this",response)
		
			if(strings.Contains(response,"ERR_REDIRECT")){
				returnedPort,_ := strconv.Atoi(strings.Fields(response)[1])
				if returnedPort != -1 {
					conn.Close()
					return strings.Fields(response)[1]
				}
			}else {
				conn.Close()
				return tempId
			}
		}else{
			//fmt.Println("error+",err.Error())
		}
	}
	return "-1"

}

func TestMain(m *testing.M){
	//go serverMain()
	//time.Sleep(1*time.Second)
	//0. Cleanup
	removeData()
	fmt.Println("hi")
	var args[] string=make([]string,3)
	args[1]="1"
	args[2]=os.ExpandEnv(path_to_config)
	//args[1] = "config"
	path_to_binary_exp := os.ExpandEnv(path_to_binary)
	for i:=0;i<total;i++{
		args[1] = strconv.Itoa(i+1)
		p,err := os.StartProcess(path_to_binary_exp,args, &os.ProcAttr{Files: []*os.File{nil, os.Stdout, os.Stderr}})
		servers = append(servers,p)
		if err != nil{
			fmt.Println("Error+",err.Error())
			return 
		}
	}
	time.Sleep(7*time.Second)
	
	m.Run()
	//fmt.Println("Killing all the servers")
	for i:=0;i<len(servers);i++{
		servers[i].Kill()
	}
	
}

func TestProcess(t *testing.T){
	
	//test the working of cluster
	//leaderPort := findLeaderPort([]int{5000,5001,5002,5003,5004,5005})	
	leaderPort := findLeaderPort(ports)
	//fmt.Println(leaderPort)
	url := "127.0.0.1:"+leaderPort
	conn, _ := net.Dial("tcp",url)
	connReader := bufio.NewReader(conn)

	conn.Write([]byte("write myfile 10 10000\r\nabc\x00\xFF\x011234\r\n"))
	response,_ := connReader.ReadString('\n')
	//fmt.Println(response)
	if (!strings.Contains(response,"OK")){
		fmt.Println(response)
		t.Errorf("Error")
	}
	
}

func TestLeaderFailure(t *testing.T){
	
	//write to a file
	//leaderPort := findLeaderPort([]int{5000,5001,5002,5003,5004})	
	leaderPort := findLeaderPort(ports)
	//fmt.Println(leaderPort)
	url := "127.0.0.1:"+leaderPort
	conn, _ := net.Dial("tcp",url)
	connReader := bufio.NewReader(conn)

	conn.Write([]byte("write myfile1 2 10000\r\nhi\r\n"))
	response,_ := connReader.ReadString('\n')
	//fmt.Println(response)
	if (!strings.Contains(response,"OK")){
		fmt.Println(response)
		t.Errorf("Error")
	}
	time.Sleep(1*time.Second)
	//kill the leader process
	servers[0].Kill()
	
	//let the new leader be elected
	time.Sleep(7*time.Second)

	leaderPort = findLeaderPort([]int{5000,5001,5002,5003,5004,5005})	
	//leaderPort = findLeaderPort(ports)	
	//fmt.Println(leaderPort)
	url = "127.0.0.1:"+leaderPort
	conn, _ = net.Dial("tcp",url)
	connReader = bufio.NewReader(conn)

	conn.Write([]byte("read myfile1\r\n"))
	response,_ = connReader.ReadString('\n')
	contents,_ := connReader.ReadString('\n') //Beware it might get hanged here
	//fmt.Println(string(contents))
	if (!strings.Contains(response,"CONTENTS") || !strings.Contains(contents,"hi")){
		fmt.Println(string(contents))
		t.Errorf("Error")
	}
	
	conn.Write([]byte("write myfile1 6 10000\r\njayesh\r\n"))
	response,_ = connReader.ReadString('\n')
	//fmt.Println(response)
	if (!strings.Contains(response,"OK")){
		fmt.Println(response)
		t.Errorf("Error")
	}
	
	conn.Write([]byte("read myfile1\r\n"))
	response,_ = connReader.ReadString('\n')
	contents,_ = connReader.ReadString('\n') //Beware it might get hanged here
	//fmt.Println(string(contents))
	if (!strings.Contains(response,"CONTENTS") || !strings.Contains(contents,"jayesh")){
		fmt.Println(string(contents))
		t.Errorf("Error")
	}
}

func TestFollowerFailure(t *testing.T){
	//write to a file
	//leaderPort := findLeaderPort([]int{5000,5001,5002,5003,5004})	
	leaderPort := findLeaderPort(ports)
	//fmt.Println(leaderPort)
	url := "127.0.0.1:"+leaderPort
	conn, _ := net.Dial("tcp",url)
	connReader := bufio.NewReader(conn)

	conn.Write([]byte("write myfile2 2 10000\r\nhi\r\n"))
	response,_ := connReader.ReadString('\n')
	//fmt.Println(response)
	if (!strings.Contains(response,"OK")){
		fmt.Println(response)
		t.Errorf("Error")
	}
	time.Sleep(1*time.Second)
	//kill the follower process
	serverToFail := 2
	servers[serverToFail].Kill()
	
	//let the new leader be elected
	//time.Sleep(7*time.Second)

	conn.Write([]byte("read myfile2\r\n"))
	response,_ = connReader.ReadString('\n')
	contents,_ := connReader.ReadString('\n') //Beware it might get hanged here
	//fmt.Println(string(contents))
	if (!strings.Contains(response,"CONTENTS") || !strings.Contains(contents,"hi")){
		fmt.Println(string(contents))
		t.Errorf("Error")
	}
	
	conn.Write([]byte("write myfile2 6 10000\r\njayesh\r\n"))
	response,_ = connReader.ReadString('\n')
	//fmt.Println(response)
	if (!strings.Contains(response,"OK")){
		fmt.Println(response)
		t.Errorf("Error")
	}
	
	conn.Write([]byte("read myfile2\r\n"))
	response,_ = connReader.ReadString('\n')
	contents,_ = connReader.ReadString('\n') //Beware it might get hanged here
	//fmt.Println(string(contents))
	if (!strings.Contains(response,"CONTENTS") || !strings.Contains(contents,"jayesh")){
		fmt.Println(string(contents))
		t.Errorf("Error")
	}
	
	var args[] string=make([]string,3)
	fmt.Println("here")
	args[1]="3"
	//args[2]="/home/jayesh/go_work/src/github.com/jayeshk/cs733/config"
	args[2]=os.ExpandEnv(path_to_config)
	
	p,err := os.StartProcess(os.ExpandEnv(path_to_binary),args, &os.ProcAttr{Files: []*os.File{nil, os.Stdout, os.Stderr}})
	if(err == nil){
		//fmt.Println("success")
		servers[serverToFail]=p
	}
	conn.Write([]byte("write myfile2 6 10000\r\nrajesh\r\n"))
	response,_ = connReader.ReadString('\n')
	
	//let the follower be upto date
	time.Sleep(5*time.Second)
	
	//make this follower the leader and check if working fine
	servers[0].Kill()
	servers[1].Kill()
	
	//let the election happen
	time.Sleep(7*time.Second)
	
//	leaderPort = findLeaderPort([]int{5000,5001,5002,5003,5004,5005})	
	leaderPort = findLeaderPort(ports)
	fmt.Println(leaderPort)
	url = "127.0.0.1:"+leaderPort
	conn, _ = net.Dial("tcp",url)
	connReader = bufio.NewReader(conn)

	conn.Write([]byte("read myfile2\r\n"))
	response,_ = connReader.ReadString('\n')
	contents,_ = connReader.ReadString('\n') //Beware it might get hanged here
	//fmt.Println(string(contents))
	if (!strings.Contains(response,"CONTENTS") || !strings.Contains(contents,"rajesh")){
		fmt.Println(string(contents))
		t.Errorf("Error")
	}
	
}
