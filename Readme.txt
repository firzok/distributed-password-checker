
Requirements:
  Go Installed

Steps:
  1. Create a folder named Password and place the password file in it. (file should be named passwords.txt)

  2. Use the splitter.go file to split the files in folder passwordSplitFiles. (go run splitter.go)

  3. On slave machines create different slave folders i.e (slave1, slave2, slave3 etc)

  4. Under slave folders create a folder named passwordSplitFiles.

  5. Inside different slaves under the passwordSplitFiles folder place random split files (generated in step 2).

  6. Now run the server. (go run server.go)
    you can provide the following flags:
    -clientPort  default:8000 :::: Port on which server will listen for client connection.
    -slavePort   default:8001 :::: Port on which server will listen for slave connection.
    -localClient default:true :::: True if client is local, false otherwise.

  7. Run client. (go run client.go)
    you can provide the following flags:
    -clientPort  default:9090 :::: Port on which client will run on localhost.
  	-serverPort  default:8000 :::: Port on which client will connect to server.
  	-serverIP    default:127.0.0.1 :::: IP on which client will connect to server.
    
  8. Run slaves. (go run slave.go -serverIP=(serverIP provided by the server))
    you can provide the following flags:
    -serverPort default:8001  :::: Port on which slave will connect to server
  	-serverIP default:127.0.0.1 :::: IP on which slave will connect to server

  9. Go to the link provided on client and search for passwords.
