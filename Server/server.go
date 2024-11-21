package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type message struct {
	Author       string
	Content      string
	Timestamp    string
	mood         bool
	TargetedRoom string // name of the room
	Kind         string // The message either for communication purpose (Plaint message) or a message to trigger something
}

type client struct {
	NickName string
	Online   bool
}

type room struct {
	Name    string
	Clients []*client
}

var AvailableRooms []*room

var upgrader = websocket.Upgrader{}

var AllClients []*client

func EstablishUpgradedConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return c, fmt.Errorf("failed to upgrade the connection !! ; error : %s ", err)
	}
	_, msg, err := c.ReadMessage()
	if err != nil {
		return c, fmt.Errorf("error occured while reading byte of received message")
	}
	var MSG message
	err = json.Unmarshal(msg, &MSG)
	if err != nil {

		return c, fmt.Errorf("failed to unmarshal the received data : %s", err)

	}
	if MSG.Kind == "init" {
		nickname := MSG.Author
		Response := message{
			Author:       "",
			Content:      "",
			Timestamp:    " ",
			mood:         false,
			TargetedRoom: "",
			Kind:         "init",
		}
		for _, name := range AllClients {
			if nickname == name.NickName {
				Response.Content = "Duplicated Nickname , Nickname provoided already existed in our database"
				ActualMSG, err := json.Marshal(Response)
				if err != nil {

					return c, fmt.Errorf("error while marshalling response to the client :%s ", err)

				}

				err = c.WriteMessage(websocket.TextMessage, ActualMSG)
				if err != nil {
					return c, fmt.Errorf("error occured while trying to send back the message of SIgnedUp to the client")
				}
			}
		}
		Response.mood = true
		ActualMSG, err := json.Marshal(Response)
		if err != nil {
			return c, fmt.Errorf("error while marshalling response to the client : %s ", err)
		}
		err = c.WriteMessage(websocket.TextMessage, ActualMSG)
		if err != nil {
			return c, fmt.Errorf("error occured while trying to send back the message of signUp to the client")
		}

	} else {
		return c, fmt.Errorf("the server expcted a message of type init but kind %s is provided", MSG.Kind)
	}
	return c, nil
}

func JoinClientRoom(c *websocket.Conn) {
	_, msg, err := c.ReadMessage()
	if err != nil {
		fmt.Println("Error occured while reading the message , error : ", err)
		return
	}
	var Response message
	err = json.Unmarshal(msg, &Response)
	if err != nil {
		fmt.Println("Error while unmarshalling the message !!") // In this case the client will wait for the response indefintely , we assume that won't happen in beta version
		return
	}
	if Response.Kind == "negotiate_name" { // means it's a new room
		response_msg := message{
			Author:       "",
			Content:      "",
			Timestamp:    "",
			mood:         true,
			TargetedRoom: "", // name of the room
			Kind:         "negotiate_name",
		}
		for _, room := range AvailableRooms {
			if Response.Content == room.Name {
				response_msg.mood = false // Server not ok
				c.WriteMessage()
				return
			}
		}
	}

}
