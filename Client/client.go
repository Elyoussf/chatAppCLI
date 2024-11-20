package client

import (
	"encoding/json"
	"fmt"

	"time"

	"github.com/gorilla/websocket"
)

var NickName string

type message struct {
	Author       string
	Content      string
	Timestamp    string
	mood         bool
	TargetedRoom string // name of the room
	Kind         string // The message either for communication purpose (Plaint message) or a message to trigger something
}

// important thing : a caller to this should end the Connection
func ConnecToTheServer() (string, error, *websocket.Conn) {

	Conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		fmt.Println("Failed to Connect to the server ")
		return "", err, nil
	}
	for {

		var Author string
		fmt.Println("Please write your nickname")
		fmt.Scan(Author)

		SignUpUser, err := json.Marshal(message{
			Author:       Author,
			Content:      "",
			Timestamp:    "",
			mood:         true,
			TargetedRoom: "",
			Kind:         "init",
		})

		if err != nil {
			fmt.Println("Failed to marshal the data")
			return "", err, Conn
		}

		err = Conn.WriteMessage(websocket.TextMessage, SignUpUser)

		if err != nil {
			fmt.Println("The Client Connected successfully ,However could not be registered successfully !! ")
			return "", err, Conn
		}
		fmt.Println("Waiting for The server to check its database")

		_, msg, err := Conn.ReadMessage()
		var Response message
		err = json.Unmarshal(msg, &Response)
		if err != nil {
			fmt.Println("Failed to unmarshal the data received from the server while init phase")
			return "", err, Conn
		}
		if Response.Kind == "init" {
			if !Response.mood {
				fmt.Println("The server was not ok within init phase and he responded with this error: ", Response.Content)
				time.Sleep(time.Second)
				fmt.Println("let's start over!! ")
			} else {
				fmt.Println("Signed up successfully")
				NickName = Author
				return Author, nil, Conn
			}
		} else {
			return "", fmt.Errorf("The server responding out of context , Contact Hamza"), Conn
		}
	}
}

func DetermineMyRoom(Author string, Conn *websocket.Conn) string { // returns the name of the room
	var ClientName string
	// The room is either an existing one or the system is prompted to create a new one
	// we gonna rely on a ui that will interact with the user in this case
	// At this point we suppose there is a procedure that does that and returns the user choice
	var ExistingRoom bool // This is true if he has choosen an existing one Otherwise false
	if !ExistingRoom {
		// we gotta create a new room with the given name "ClientName"
		NewRoom, err := json.Marshal(message{
			Author:       Author,
			Content:      ClientName, // send the name of the room
			Timestamp:    "",
			mood:         true,
			TargetedRoom: "", // This wil be used during the chat
			Kind:         "new_room",
		})

		if err != nil {
			fmt.Println("Error while marshalling the json ")
			return ""
		}

		err = Conn.WriteMessage(websocket.TextMessage, NewRoom)

		if err != nil {
			fmt.Println("Ask for a room Operation failed")
			return ""
		}

		_, msg, err := Conn.ReadMessage()
		var Response message
		err = json.Unmarshal(msg, &Response)
		if err != nil {
			fmt.Println("Failed to unmarshal the data received from the server while init phase")
			return ""
		}

		if Response.Kind == "new_room" {
			if Response.mood {
				fmt.Println("You room created successfully")
				return Response.Content
			} else {
				fmt.Println("The Server could not create the room for some reason , it specified:  ", Response.Content)
			}
		} else {
			fmt.Println("The Server responding in out of context hehe")
			return ""
		}
		return ""
	} else {
		SpecifyRoom, err := json.Marshal(message{
			Author:       Author,
			Content:      ClientName, // send the name of the room
			Timestamp:    "",
			mood:         true,
			TargetedRoom: "", // This wil be used during the chat
			Kind:         "add_to_room",
		})

		if err != nil {
			fmt.Println("Error while marshalling the json ")
			return ""
		}

		err = Conn.WriteMessage(websocket.TextMessage, SpecifyRoom)

		if err != nil {
			fmt.Println("Ask for a room join has been failed")
			return ""
		}
		_, msg, err := Conn.ReadMessage()
		var Response message
		err = json.Unmarshal(msg, &Response)
		if err != nil {
			fmt.Println("Failed to unmarshal the data received from the server while init phase")
			return ""
		}
		if Response.Kind == "add_to_room" {
			if Response.mood {
				fmt.Println("You have been added successfully to the room ", Response.Content)
				return Response.Content
			} else {
				fmt.Println("The join operation failed ")
				return ""
			}
		} else {
			fmt.Println("The bug in server !! , contact El youssfi")
			return ""
		}

	}

}

func askForUserChoice(Conn *websocket.Conn) (string, bool) { // string is for the name of the room , the boolean value tells whether it is an existing room or an invented one

	var number string

	for number != "1" && number != "2" {
		fmt.Println("You Want to  create a new room or join an existing one ? (choose a number) ")
		fmt.Println("1 - Create a new room \n 2 - Join an existing one ")
		fmt.Scan(number)

	}
	if number == "1" {
		for {
			var name string
			fmt.Println("What is the prefered name ")
			fmt.Scan(name)
			NegotiateName, err := json.Marshal(message{
				Author:       "",
				Content:      name, // send the name of the room
				Timestamp:    "",
				mood:         true,
				TargetedRoom: "", // This wil be used during the chat
				Kind:         "negotiate_name",
			})
			if err != nil {
				fmt.Println("error while marshalling data, please start over ")
				fmt.Println("let's start over")
				continue
			}
			err = Conn.WriteMessage(websocket.TextMessage, NegotiateName)
			if err != nil {
				fmt.Println("Error occured while sending the message ")
				fmt.Println("let's start over")
				continue
			}
			_, msg, err := Conn.ReadMessage()
			if err != nil {
				fmt.Println("an error occured while decoding the received message")
				fmt.Println("will resend it ")
				continue
			}
			var Response message
			err = json.Unmarshal(msg, &Response)
			if err != nil {
				fmt.Println("error to unmarshall the message")
				fmt.Println("will start over")
				continue
			}
			if Response.Kind == "negotiate_name" {
				if Response.mood {
					fmt.Println("You have been added successfully to the room ")
					time.Sleep(50 * time.Millisecond)
					return name, false
				} else {
					fmt.Println("The server was not Okay and justified with : ", Response.Content)
					fmt.Println("We will start over")
					continue
				}
			} else {
				fmt.Println("The server was not expected to answer in this way , contact hamza")
			}
		}

	} else {
		// Send to the server to give you the list of the available room to choose from!! ()
	}
}
