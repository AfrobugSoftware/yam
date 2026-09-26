package main

import (
	"fmt"
	"log"
	"yam/ygame"
)

func main() {
	g, err := ygame.NewGame("Kiwi Island", 1000, 600)
	if err != nil {
		log.Panicf("Error creating game: %v", err)
	}
	fmt.Println("Game created successfully!")
	g.SetApplication(&ygame.TestApplication{})
	g.Run()
}
