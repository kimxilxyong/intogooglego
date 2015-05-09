package main

import (
	"fmt"
	"github.com/kimxilxyong/rpcbotinterfaceobjects"
	"log"
	"net/rpc"
)

func main() {

	client, err := rpc.Dial("tcp", "localhost:9876")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	// Synchronous call
	in := rpcbotinterfaceobjects.BotInput{"BotInputTestText"}
	var out rpcbotinterfaceobjects.BotOutput

	err = client.Call("Bot.ProcessPost", in, &out)
	if err != nil {
		log.Fatal("ProcessPost error:", err)
	}
	fmt.Printf("ProcessPost: %s\n", out.GetContent())
}
