* This repository contains implementation of replicated, consistent, in-memory file system which can handle large number of concurrent clients. Such system generally consists of several servers deployed at various locations forming a tight cluster, such that each server in cluster contains consistent replica of the file system. A client of the file system can contact to any of these servers, and this server redirects client to current leader in the cluster. Client can then communicate to leader to perform it's required operations.

* The main challenge here is to maintain consistent version of replicas on each server in cluster. For this our system uses RAFT distributed consensus protocol. For details of RAFT, refer to:

* Specifically this repository contains the implementation of one server in such cluster. When initializing, a configuration of cluster must be provided to each server so that it can communicate to it's peers.

* Usage:

- Setup
Assuming you have proper setup of golang build system.
1) Pull the project located at https://github.com/jayeshjk/CS733Project
2) cd $GOPATH/src/github.com/jayeshk/cs733/assignment4
3) go build
4) go install

This will create a binary of assignment4 in $GOPATH/bin.

- To run:

1) Create a config file. Format of config file is intuitive. It contains details of each server in cluster. Each server process listens on two different ports, one to serve the clients and other to communicate with it's peers in cluster. So for each server in cluster, config file contains id, clientPort, raftPort. Hence format of config file is:
id1
clientPort1
raftPort1
id2
clientPort2
raftPort2
... and so on.

A sample config file is provided in repository.

2) Suppose server binary created during setup is located at $GOPATH/bin/assignment4. Then use following command to start the server.
$GOPATH/bin/assignment4 server-id path-to-config-file

Here the server-id must be one of the ids in config file supplied as second argument.

3) Start all the servers specified in config file, and wait for few seconds(2-3 seconds), so that servers find each other, and agree upon some parameters. Now, you can contact to any server in the cluster.

- To Test:
1) You should have first built and installed the assignment4 binary, because test cases use that binary to start the servers autmatically.
2) Further, testcases assume that binary is located at $GOPATH/bin/assignment4 and config file at $GOPATH/src/github.com/jayeshk/cs733/assignment4/config. Change the variables "path_to_config" and "path_to_binary" in server_test.go if your config and binary are located at different locations.
3) Now do go test.
4) Test takes around 150 seconds to finish. You should see PASS and OK in the output.

* File server protocol:

1. Write: create a file, or update the file’s contents if it already exists. Command is:
   
   write filename numbytes [exptime]\r\n
   content bytes\r\n

   The server responds with the following:
   OK version\r\n
   where version is a unique 64‑bit number (in decimal format) assosciated with the
   filename.

2. Read: Given a filename, retrieve the corresponding file:

   read filename\r\n
   
   The server responds with the following format:
   CONTENTS version numbytes exptime \r\n
   content bytes\r\n

3. Compare and swap. This replaces the old file contents with the new content provided the version is still the same.
   
   cas filename version numbytes [exptime]\r\n
   content bytes\r\n

   The server responds with the new version if successful
   OK version\r\n

4. Delete file

   delete filename\r\n
   Server response (if successful)
   OK\r\n


* Various options in above commands are:
1) filename : A text string without spaces.
2) numbytes: No of bytes, an int.
3) exptime : After exptime seconds, the file will be deleted at the server, an int.
4) version: A 64 bit int.

* Various Error messages are as follows:

1) "ERR_REDIRECT leader-port" secifies the port on which current leader of the cluster listens. -1 is returned if no server doesn't know about the leader. Connection to current server is closed.

2) Whenever there is error in first line of command, "ERR_CMD_ERR" is returned, and connection is closed. Then the connection will have to be reset to talk to server again. This is to ensure that formatting error in one ommand doesn't mess up processing of later commands.

3) If file is not found on the file-store at server, "ERR_FILE_NOT_FOUND" is returned. However the connection is not closed.

4) If, in case of CAS, the version mismatch occurs, "ERR_VERSION_ERR current version" is returned. And connection is not closed.
