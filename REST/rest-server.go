package main

import (
	"github.com/gorilla/mux"
	"fmt"
	"net/http"
	"net/url"
	"bytes"
	"io/ioutil"
	"github.com/rs/xid"

)

var helpInfo = [...]string{
	"help and command info:",
	"/help: use this command to get some help",
	"/createroom roomName : creates a room with the name roomName",
	"/listrooms: lists all rooms available for joining",
	"/join roomName: adds you to a chatroom",
	"/leave removes you from current room",
}
var Members []*Client
var RoomArray []*Room
var sender *Client
var member *Client
var roomToJoin *Room

type Room struct {
	name        string
	clientList  []*Client
	previousMSG []*Chat
}

type Chat struct {
	client  *Client
	message string
}

type Client struct {
	currentRoom   *Room
	outputChannel chan string
	name          string
}

func main() {

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", CreateClient).Methods("GET")

	router.HandleFunc("/help", help).Methods("GET")

	router.HandleFunc("/{USER}/messages", MessageClient).Methods("GET")

	router.HandleFunc("/rooms", listrooms).Methods("GET")

	router.HandleFunc("/rooms/{ROOMNAME}", create).Methods("POST")

	router.HandleFunc("/rooms/{ROOMNAME}/{USER}", join).Methods("POST")

	router.HandleFunc("/{USER}/leaveroom", leaveroom).Methods("DELETE")

	router.HandleFunc("/{USER}/messageRoom", Broadcast).Methods("POST")

	http.ListenAndServe(":8080", router)

}

func (cli *Client) leave() {
	//not in a current room so just return
	if cli.currentRoom == nil {
		return
	} else {
		cli.outputChannel <- "You left!"
		cl := cli.currentRoom.clientList
		for i, roomClients := range cl {
			if cli == roomClients {
				cli.currentRoom.clientList = append(cl[:i], cl[i+1:]...) //deletes the element
			}
		}
		cli.currentRoom = nil
		return
	}
}

func Broadcast(w http.ResponseWriter, r *http.Request) {

	clientName := mux.Vars(r)["USER"]
	message := r.Header.Get("message")

	for _, cli := range Members {
		if cli.name == clientName {
			sender = cli
		}
	}

	if sender.currentRoom == nil {
		sender.outputChannel <- "You are not in a room yet\n"
		return
	}

	room := sender.currentRoom

	chatMessage := &Chat{
		client:  sender,
		message: message,
	}

	for _, roomUser := range room.clientList {
		//check to see if the user is currently active in the room
		if roomUser.currentRoom.name == room.name {
			go roomUser.broadcastMessage(chatMessage.message, chatMessage.client)

		}
	}
	room.previousMSG = append(room.previousMSG, chatMessage)
	sendTBC(message,sender)
}

func (cli *Client) broadcastMessage(message string, sender *Client) {
	cli.outputChannel <- (sender.name) + ": " + message + "\n"
}

func sendTBC(message string, sender *Client) {
	request_url := "http://localhost:3000/api/ChatMessage"
	guid := xid.New()

	form := url.Values{
		"chatMessageContent": {message},
		"sender":             {sender.name},
		"room":               {sender.currentRoom.name},
		"chatMessageId":      {guid.String()},
	}

	body := bytes.NewBufferString(form.Encode())
	rsp, err := http.Post(request_url, "application/x-www-form-urlencoded", body)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	body_byte, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body_byte))

}

func CreateClient(w http.ResponseWriter, r *http.Request) {

	clientName := r.Header.Get("username")
	fmt.Println(clientName, "---------")
	cli := &Client{
		currentRoom:   nil, //starts as nil because the user is not initally in a room
		outputChannel: make(chan string),
		name:          clientName,
	}
	Members = append(Members, cli)

	reply := cli.name

	fmt.Fprintf(w, reply)
}

func MessageClient(w http.ResponseWriter, r *http.Request) {
	clientName := mux.Vars(r)["USER"]
	for _, cli := range Members {
		if cli.name == clientName {
			sender = cli
		}
	}
	reply := <-sender.outputChannel
	fmt.Fprintf(w, reply)
}

//creates a room
func create(w http.ResponseWriter, r *http.Request) {

	roomName := mux.Vars(r)["ROOMNAME"]
	clientName := r.Header.Get("username")
	for _, cli := range Members {
		if cli.name == clientName {
			member = cli
		}
	}
	//Is the name unique?
	for _, room := range RoomArray {
		if roomName == room.name {
			member.outputChannel <- "The room name you have specified is already in use\n"
			return
		}
	}
	newRoom := &Room{
		name:        roomName,
		clientList:  make([]*Client, 0), //room will start empty
		previousMSG: nil,
	}
	RoomArray = append(RoomArray, newRoom)

	member.outputChannel <- member.name + " created a new room called " + roomName + "\n"
}
func leaveroom(w http.ResponseWriter, r *http.Request) {
	clientName := mux.Vars(r)["USER"]
	for _, cli := range Members {
		if cli.name == clientName {
			member = cli
		}
	}

	member.leave()
	member.outputChannel <- "You have left the room\n"
}

func help(w http.ResponseWriter, r *http.Request) {

	clientName := r.Header.Get("username")
	fmt.Println(clientName, "help menu")
	client := &Client{
		currentRoom:   nil, //starts as nil because the user is not initally in a room
		outputChannel: make(chan string),
		name:          clientName,
	}
	for _, cli := range Members {
		if cli.name == clientName {
			client = cli
		}
	}
	for _, helpLine := range helpInfo {
		client.outputChannel <- helpLine + "\n"
	}
}

//sends the list of rooms to the client
func listrooms(w http.ResponseWriter, r *http.Request) {
	clientName := r.Header.Get("username")
	for _, cli := range Members {
		if cli.name == clientName {
			member = cli
		}
	}
	member.outputChannel <- "List of rooms:\n"
	for _, roomName := range RoomArray {
		member.outputChannel <- roomName.name + "\n"
	}
}

//returns true of the room was joined successfully, returns false if there was a problem like the room does not exist
func join(w http.ResponseWriter, r *http.Request) {
	clientName := mux.Vars(r)["USER"]
	roomName := mux.Vars(r)["ROOMNAME"]
	for _, cli := range Members {
		if cli.name == clientName {
			member = cli
		}
	}
	//checks to see if a room with the given name exists in the RoomArray, if it does return it, if not return nil
	for _, room := range RoomArray {
		if room.name == roomName {
			roomToJoin = room
		}
	}
	if roomToJoin == nil { //the room doesnt exist
		member.outputChannel <- "The room " + roomName + " does not exist\n"
		return
	}
	//add user to room
	if roomToJoin.isInRoom(member) {
		member.outputChannel <- "You are already in that room\n"
	} else { //join room and display all the rooms messages
		member.leave()
		roomToJoin.clientList = append(roomToJoin.clientList, member) // add client to the rooms list
		//switch users current room to room
		member.currentRoom = roomToJoin
		member.outputChannel <- "You are now joined!\n"
		//display all messages
		if roomToJoin.previousMSG == nil {
			return
		}
		for _, messages := range roomToJoin.previousMSG {
			member.outputChannel <- (messages.client.name) + ": " + messages.message + "\n"

		}
		member.outputChannel <- "*************\n"
	}
}

func (room Room) isInRoom(client *Client) bool {
	for _, roomClient := range room.clientList {
		if client.name == roomClient.name {
			return true
		}
	}
	return false
}
