package main

import (
	"github.com/hawkingrei/G53/servers"
	"os"
)

func main() {
	servers.StartServer(os.Args[1:])
}
