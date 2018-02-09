package main

import (
	"net"
	"bufio"
	"strings"
	"io"
	"log"
	//"strconv"
)

type client struct {
	Name    string
	Room    string
	Conn    net.Conn
	Message chan string
}

type chatroom struct {
	name     string
	members  map[string]*client
	messages chan string
}

var help = map[string]string{
	"create":    "create a new room\n",
	"join":      "join a room\n",
	"listrooms": "list all online users\n",
	"help":      "prints all available commands\n",
	"quit":      "quit\n",
}

var roomList = map[string]*chatroom{}

func main() {
	listener, err := net.Listen("tcp", ":8181")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	for {
		//connection
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		//conn.SetDeadline(time.Now().Add(7 * 24))
		go func(conn net.Conn) {

			conn.Write([]byte("What is your name?"))
			name, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				panic(err)
			}
			//name
			name = strings.Trim(name, "\r\n")

			//io.WriteString(conn, "Welcome: ")
			//io.WriteString(conn, name)
			//io.WriteString(conn, "\n")

			client := &client{
				Name:    name,
				Room:    "",
				Conn:    conn,
				Message: make(chan string),
			}

			log.Printf("New client created: %v %v", client.Conn.RemoteAddr(), client.Name)

			//help menu
			for k, v := range help {
				_, err = conn.Write([]byte(k + " : " + v))
			}

			go func() {
			Loop:
				for {
					msg, err := bufio.NewReader(conn).ReadString('\n')
					if err != nil {
						panic(err)
					}
					msg = strings.Trim(msg, "\r\n")

					if msg == "quit" {
						io.WriteString(client.Conn, "got here in first if loop")
						if client.Room != "" {
							io.WriteString(client.Conn, "got here in emty room")

							delete(roomList[client.Room].members, client.Conn.RemoteAddr().String())
							io.WriteString(client.Conn, "got here in delete")
							log.Printf("leave: removing user %v from room %v: current members: %v", client.Name, client.Room, roomList[client.Room].members)
							io.WriteString(client.Conn, "leaving "+client.Room+"\n")
						}
						client.Conn.Close()
						client.Message <- "quit"
						log.Printf("%v has left..", client.Name)
						break Loop
					}
					if client.order(msg) {
						log.Printf("send: msg: %v from: %s", msg, client.Name)
						send := client.Name + ": " + msg
						for _, v := range roomList {
							for k := range v.members {
								if k == client.Conn.RemoteAddr().String() {
									v.messages <- send
								}
							}
						}
					}
				}
			}()

			go func() {
				for {
					msg := <-client.Message
					if msg == "quit" {
						break
					}
					log.Printf("recieve: client(%v) recvd msg: %s ", client.Conn, msg)
					io.WriteString(client.Conn, msg)
					io.WriteString(client.Conn, "\n")
				}
			}()

		}(conn)

	}
}

//func (receiver) identifier(parameters) (returns) { code }
func (c client) order(msg string) bool {

	switch {
	case msg == "join":
		c.join()
		return false
	case msg == "create":
		c.create()
		return false
	case msg == "help":
		for k, v := range help {
			_, err := c.Conn.Write([]byte(k + ": " + v))
			if err != nil {
				panic(err)
			}
		}
		return false
	case msg == "listrooms":
		c.Conn.Write([]byte("Created rooms are: " + "\n"))
		for k := range roomList {
			c.Conn.Write([]byte(roomList[k].name + "\n"))

		}
		return false
	}
	return true
}

func (c *client) join() {
	c.Conn.Write([]byte("Please enter room name: "))
	roomName, err := bufio.NewReader(c.Conn).ReadString('\n')

	if err != nil {
		panic(err)
	}
	roomName = strings.Trim(roomName, "\r\n")
	//adding client
	if room := roomList[roomName]; room != nil {
		room.members[c.Conn.RemoteAddr().String()] = c
		//if c.Room != "" {
		//
		//	delete(roomList[c.Room].members, c.Conn.RemoteAddr().String())
		//	log.Printf("leave: removing user %v from room %v: current members: %v", c.Name, c.Room, roomList[c.Room].members)
		//	io.WriteString(c.Conn, "leaving "+c.Room+"\n")
		//	room.messages <- "* " + (c.Name + " has left..") + " *"
		//}
		c.Room = roomName
		io.WriteString(c.Conn, c.Name+" has joined "+room.name+"\n")
		room.messages <- "* " + (c.Name + " has joined!") + " *"
	} else {
		io.WriteString(c.Conn, "error: could not join room\n")
	}
}

func (c *client) create() {
	c.Conn.Write([]byte("Please enter room name: "))
	roomName, err := bufio.NewReader(c.Conn).ReadString('\n')
	if err != nil {
		log.Printf("readinput: could not read input from stdin: %v from client %v", err, c.Conn.RemoteAddr().String())
	}
	roomName = strings.Trim(roomName, "\r\n")
	if roomName != "" {
		room := createRoom(roomName)
		room.members[c.Conn.RemoteAddr().String()] = c
		if c.Room != "" {

			delete(roomList[c.Room].members, c.Conn.RemoteAddr().String())
			log.Printf("leave: removing user %v from room %v: current members: %v", c.Name, c.Room, roomList[c.Room].members)
			io.WriteString(c.Conn, "leaving "+c.Room+"\n")
			room.messages <- "* " + (c.Name + " has left..") + " *"
		}
		c.Room = room.name
		roomList[room.name] = room
		room.messages <- "* " + (c.Name + " has joined!") + " *"

		io.WriteString(c.Conn, "* room "+room.name+" has been created *\n")
	} else {
		io.WriteString(c.Conn, "* error: could not create room \""+roomName+"\" *\n")
	}
}

func createRoom(name string) *chatroom {
	c := &chatroom{
		name:     name,
		messages: make(chan string),
		members:  make(map[string]*client, 0),
	}
	log.Printf("creating room %v", c.name)
	//listen for messages
	go func(c *chatroom) {
		for {
			out := <-c.messages
			for _, v := range c.members {
				v.Message <- out
				log.Printf("createroom: broadcasting msg in room: %v to member: %v", c.name, v.Name)
			}
		}
	}(c)
	return c
}
