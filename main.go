package main

import (
	"fmt"
	"yam/ygame"
)

func main() {
	fmt.Println("kiwi island 1.0.2!")
	g, error := ygame.NewGame("Kiwi Island", 1000, 600)
	if error != nil {
		fmt.Println("Error creating game:", error)
		return
	}
	fmt.Println("Game created successfully!")
	g.Run()
}
