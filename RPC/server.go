package main

import (
	"net"
	"log"
	"net/rpc"
	"net/http"
	"strings"
)

type Server struct {
	rooms   map[string]Chatroom
	clients map[string]Client
	output  chan string
}

type Client struct {
	name     string
	messages []string
}

type Chatroom struct {
	name     string
	clients  []string
	messages []string
}

func main() {

	cs := Server{
		rooms:   make(map[string]Chatroom),
		clients: make(map[string]Client),
	}

	rpc.Register(&cs)
	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", ":9999")

	if err != nil {
		log.Panic("Can't bind port to listen.")
	}

	http.Serve(listener, nil)
}

func (c *Server) Register(username string, reply *string) error {
	log.Printf("Setting User: %v", username)
	client := Client{
		name: username,
	}
	c.clients[client.name] = client
	*reply = "Your user name has been created. Choose from options below!"
	return nil
}

func (c *Server) Create(name string, reply *string) error {

	c.rooms[name] = Chatroom{name, nil, nil}
	*reply = "chatRoom was created !"
	return nil

}

func (c *Server) Leave(msg string, reply *string) error {

	user, roomName := parseMessage(msg)

	for k := range c.rooms {
		if k == roomName {
			for i := range c.rooms[k].clients { //Go through all the clients
				if c.rooms[k].clients[i] == user { //check for the specific user
					chatRoom := c.rooms[k]
					chatRoom.clients = append(chatRoom.clients[:i], chatRoom.clients[i+1:]...) //delete left clients
					c.rooms[k] = chatRoom//update chat room
					*reply = "You left the ChatRoom"
					return nil
				}
			}
		}
	}
	*reply = "You are not in the chat room"
	return nil
}

func (c *Server) List(msg string, reply *string) error {

	for room := range c.rooms {
		*reply += room
	}
	return nil
}

func (c *Server) Join(msg string, reply *string) error {

	user, roomName := parseMessage(msg)
	if _, flag := c.rooms[roomName]; flag {//to check if room name exists
		chatRoom := c.rooms[roomName]
		Clients := c.rooms[roomName].clients

		for k := range Clients {
			if Clients[k] == user {
				*reply = "You are already in this chat room"
				return nil
			}
		}
		chatRoom.clients = append(Clients, user) //update it
		c.rooms[roomName] = chatRoom

		*reply = "You joined the chat room!"

		var MessagesInRoom string//showing previous messages

		for _, msg := range c.rooms[roomName].messages {
			MessagesInRoom = MessagesInRoom + msg
		}

		*reply = *reply + "\n" + MessagesInRoom
		return nil

	} else {
		*reply = "That chat room doesn't exist"
		return nil
	}
}

func (c *Server) GetMessages(msg string, reply *string) error {

	user, _ := parseMessage(msg)

	if len(c.clients[user].messages) == 0 {
		return nil
	} else {
		var availableMessages string
		for _, msg := range c.clients[user].messages {
			availableMessages = availableMessages + msg
		}
		*reply = availableMessages

		client := c.clients[user]

		client.messages = make([]string, 0)

		c.clients[user] = client
	}

	return nil

}

func (c *Server) ReceiveMessage(msg string, reply *string) error {
	log.Printf("Message Received: %v", msg)

	go func() {

		user, message := parseMessage(msg)

		message = user + ": " + message

		Rooms := make([]string, 0)

		for room := range c.rooms {
			flag := true
			for i := range c.rooms[room].clients {
				if c.rooms[room].clients[i] == user {
					Rooms = append(Rooms, room)
					flag = false
					chatRoom := c.rooms[room]
					chatRoom.messages = append(chatRoom.messages, message)
					c.rooms[room] = chatRoom
					continue
				}
			}
			if !flag {
				continue
			}
		}

		for room := range Rooms {
			for i := range c.rooms[Rooms[room]].clients {
				if c.rooms[Rooms[room]].clients[i] != user {
					messagesToDeliver := append(c.clients[c.rooms[Rooms[room]].clients[i]].messages, message)

					client := c.clients[c.rooms[Rooms[room]].clients[i]]
					client.messages = messagesToDeliver

					c.clients[c.rooms[Rooms[room]].clients[i]] = client
				}
			}
		}
	}()

	*reply = "/n"
	return nil
}

func parseMessage(data string) (string, string) {
	result := strings.SplitN(data, "/;", 2)
	if len(result) < 2 {
		return result[0], ""
	}
	return result[0], result[1]

}
