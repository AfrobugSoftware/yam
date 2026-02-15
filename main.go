package main

import (
	"log"
	"yam/ygame"
)

const (
	height = 720
	width  = 1280
)

func main() {
	game, err := ygame.NewGame("a game", width, height)
	if err != nil {
		log.Fatal(err)
	}
	game.Run()
}
