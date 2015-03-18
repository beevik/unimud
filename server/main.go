package main

import "github.com/beevik/unimud"

func main() {
	game := unimud.NewGame()
	go game.ListenConsole()
	go game.ListenNet(2000)
	go game.ListenNet(3389)
	go game.Run()
	<-game.DoneChan
}
