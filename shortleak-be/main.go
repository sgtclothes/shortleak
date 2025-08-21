package main

import (
	"shortleak/server"
)

func main() {
	r := server.SetupRouter()
	r.Run(":8080")
}
