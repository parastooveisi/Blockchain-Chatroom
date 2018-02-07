package main

import (
	"net/rpc"
	"log"
	"fmt"
	"bufio"
	"os"
	"strings"
	//"time"
)

var userName string

func main() {
	conn, err := rpc.DialHTTP("tcp", ":9999")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	// Synchronous call
	fmt.Println("Please set your username:")
	reader := bufio.NewReader(os.Stdin)
	userName, _ = reader.ReadString('\n')

	var reply string
	err = conn.Call("Server.Register", userName, &reply)
	if err != nil {
		log.Fatal("Error:", err)
	}
	fmt.Printf("%v\n", reply)

	go getMessages(conn)

	fmt.Println("")
	fmt.Println("  create: Create a chatroom.")
	fmt.Println("  list: List chatrooms.")
	fmt.Println("  join: Join existing chatroom.")
	fmt.Println("  broadcast: Send Message to all joined chatrooms")
	fmt.Println("  leave: Quit chatroom.")
	fmt.Println("  help: Show Menu")
	fmt.Println("")

	for true {
		InputHandler(conn)
	}
}

func InputHandler(conn *rpc.Client){
	reader := bufio.NewReader(os.Stdin)

	for true {
		m, _ := reader.ReadString('\n')
		option, args := parseInput(m)
		if option==""{
			fmt.Printf("Please select an option")
		}else{
			option = strings.Replace(option,"\n","",-1)

			switch option {
			//Show the menu
			case "help":
				fmt.Println("")
				fmt.Println("  create: Create a chatroom.")
				fmt.Println("  list: List chatrooms.")
				fmt.Println("  join: Join existing chatroom.")
				fmt.Println("  broadcast: Send Message to all joined chatrooms")
				fmt.Println("  leave: Quit chatroom.")
				fmt.Println("  help: Show Menu")

			case "create":
				if args==""{
					fmt.Printf("Not enough input!")
				}else{
					var reply string
					err := conn.Call("Server.Create", args, &reply)
					fmt.Printf("%v\n", reply)
					if err != nil {
						log.Fatal("Error:", err)
					}
				}

			case "list":
				var reply string
				err := conn.Call("Server.List", args, &reply)
				if err != nil {
					log.Fatal("Error:", err)
				}
				fmt.Printf("%v", reply)

			case "join":
				if args==""{
					fmt.Println("Not enough input!")
				}else{
					var reply string
					err := conn.Call("Server.Join", userName+"/;"+args, &reply)
					fmt.Printf("%v\n", reply)
					if err != nil {
						log.Fatal("Error:", err)
					}
				}

			case "broadcast":
				if args==""{
					fmt.Println("Not enough input!")
				}else{
					var reply string
					err := conn.Call("Server.ReceiveMessage", userName+"/;"+args, &reply)
					if err != nil {
						log.Fatal("Error:", err)
					}
				}

			//Leave chatroom
			case "leave":
				if args==""{
					fmt.Println("Not enough input!")
				}else{
					var reply string
					err := conn.Call("Server.Leave", userName+"/;"+args, &reply)
					fmt.Printf("%v\n", reply)
					if err != nil {
						log.Fatal("Error:", err)
					}
				}
			}
		}
	}
}

func getMessages(conn *rpc.Client){
	for{
		var reply string
		err := conn.Call("Server.GetMessages", userName, &reply)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		if reply != "" {
			fmt.Printf("%v", reply)
		}
	}
}

func parseInput(m string)(string, string){
	splitted := strings.SplitN(m," ",2)
	if len(splitted)>1{
		return splitted[0],splitted[1]
	}
	if len(splitted)==1{
		return splitted[0],""
	}
	return "",""
}