package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
	"yam/ycore"
	"yam/ygame"

	"github.com/rs/xid"
)

func TestNetwork() {

	nm := ycore.NewNetManager("localhost", "", "5000")
	fmt.Println("Starting server")
	go nm.StartServer()

	fmt.Println("Waiting for server to start up")
	time.Sleep(5 * time.Second)
	fmt.Println("Creating client")
	cli, err := nm.CreateClient()
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(os.Stdin)
	np := <-cli.In
	id, err := xid.FromBytes(np.Data)
	if err != nil {
		panic(err)

	}
	for scanner.Scan() {
		message := scanner.Text()
		np := ycore.NetPackage{
			TY:     7,
			Sender: id,
			Data:   []byte(message),
		}
		cli.Out <- np
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	g, err := ygame.NewGame("Kiwi Island", 1000, 600)
	if err != nil {
		log.Panicf("Error creating game: %v", err)
	}
	fmt.Println("Game created successfully!")
	g.SetApplication(&ygame.TestApplication{})
	g.Run()

}
