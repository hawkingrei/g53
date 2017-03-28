package main

import (
	"github.com/hawkingrei/g53/servers"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	servers.StartServer(os.Args[1:])
}
