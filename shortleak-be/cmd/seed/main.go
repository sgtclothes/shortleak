package main

import (
	"shortleak/config"
	"shortleak/database"
)

func RunSeed() {
	cfg := config.LoadConfig()
	database.ConnectDBFunc(cfg)
	database.SeedFunc()
}

func main() {
	RunSeed()
}
