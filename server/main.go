package main

import (
	"flag"

	"github.com/beevik/unimud"
)

var (
	console bool
	port    int
)

func init() {
	flag.BoolVar(&console, "c", false, "launch with a console listener")
	flag.IntVar(&port, "port", 2000, "network listening port (use 0 for none)")
}

func main() {
	flag.Parse()

	game := unimud.NewGame()
	if console {
		go game.ListenConsole()
	}
	if port > 0 {
		go game.ListenNet(port)
	}
	go game.Run()
	<-game.DoneChan
}
