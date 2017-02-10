package main

import (
	"github.com/hawkingrei/g53/servers"
	"os"
)

func main() {
	servers.StartServer(os.Args[1:])
}
