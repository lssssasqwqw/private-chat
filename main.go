package main

import "server-client/ser"

func main() {

	server := ser.Newserver("127.0.0.1", 8888)
	server.Start()
}

// func main() {
// 	server := Newserver("127.0.0.1", 8888)
// 	server.Start()
// }
