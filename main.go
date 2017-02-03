package main

import (
	"os"
	"github.com/hawkingrei/G53/servers"
)


func main() {
        servers.StartServer(os.Args[1:])
}
