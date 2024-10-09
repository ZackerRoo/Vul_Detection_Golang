package main

import (
	"backend/client"
	"backend/server"
	"backend/web"
	"fmt"
	"os"
)

func main() {
	// 忘记了一个很重要的问题就是函数要对外面可见一定要大写

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [server|client]")
		return
	}
	switch os.Args[1] {
	case "server":
		server.ServerMain()
	case "client":
		client.ClientMain()
	case "ginweb":
		web.GinMain()
	default:
		fmt.Println("Invalid argument. Use 'server' or 'client'.")
	}

}
