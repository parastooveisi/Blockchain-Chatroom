package main

import "io/ioutil"
import "fmt"
import "bufio"
import "os"
import "strings"
import "net/http"

var flag = true

//starts up the client, starts the receiving thread and the input threads and then loops forever
func main() {

	fmt.Println("Attempting to connect to localhost: 8080")
	fmt.Println("What is your name?")
	reader := bufio.NewReader(os.Stdin)
	userName, _ := reader.ReadString('\n')
	userName = strings.TrimSpace(userName)
	fmt.Printf("Your user name is: %v", userName)
	fmt.Println("")
	client := &http.Client{
		CheckRedirect: nil,
	}
	reply, err := http.NewRequest("GET", "http://localhost:8080/", nil)
	reply.Header.Add("username", userName)
	client.Do(reply)
	if err != nil {
		fmt.Println(err)
	}

	go func() {

		for flag {
			reader := bufio.NewReader(os.Stdin)
			message, _ := reader.ReadString('\n') //read from stdin till the next newline
			var err error
			message = strings.TrimSpace(message)       //strips the newlines from the input
			command := strings.HasPrefix(message, "/") //checks to see if the line starts with /
			if command {
				//take only command part
				parsedcmd := strings.Split(message, " ")

				switch {

				case strings.HasPrefix(parsedcmd[0], "/help"):
					err = Helper("GET", "http://localhost:8080/help", userName)

				case strings.HasPrefix(parsedcmd[0], "/createroom"):
					err = Helper("POST", "http://localhost:8080/rooms/" + parsedcmd[1], userName)

				case strings.HasPrefix(parsedcmd[0], "/listrooms"):
					err = Helper("GET", "http://localhost:8080/rooms", userName)

				case strings.HasPrefix(parsedcmd[0], "/join"):
					err = Helper("POST", "http://localhost:8080/rooms/" + parsedcmd[1] + "/"+userName, userName)

				case strings.HasPrefix(parsedcmd[0], "/leave"):
					err = Helper("DELETE", "http://localhost:8080/" + userName + "/leaveroom", userName)
				}
				// it's a message
			} else if flag {
				//create a post request
				client := &http.Client{
					CheckRedirect: nil,
				}
				sendReply, _ := http.NewRequest("POST", "http://localhost:8080/" + userName + "/messageRoom", nil)
				sendReply.Header.Add("message", message)
				client.Do(sendReply)
			}
			if err != nil {
				fmt.Println(err)
			}
		}


	}()
	go func() {
		for flag {
			resp, err := http.Get("http://localhost:8080/" + userName + "/messages")
			if err != nil {
				fmt.Println("error in getting messages")
				fmt.Println(err)
				flag = false
				return
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Print(string(body))
		}
	}()
	for flag {
		//until sets to false
	}
}

//creates the http message that will be sent to the server
func Helper(method string, url string, name string) error {
	client := &http.Client{
		CheckRedirect: nil,
	}
	reply, err := http.NewRequest(method, url, nil)
	reply.Header.Add("username", name)
	client.Do(reply)
	return err
}
