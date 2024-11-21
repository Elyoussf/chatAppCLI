package main

import (
	"fmt"
	client "websocket/Client"
)

func main() {
	fmt.Println("Note : Some non valid input could break connection and you need to start over ")
	Author, err, Conn := client.ConnecToTheServer()
	defer Conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(Author) != 0 && Conn != nil {
		RoomName := client.DetermineMyRoom(Author, Conn)
		for {
			go client.SendMessage(Conn, Author, RoomName)
			go client.ReceiveMessage(Conn)
		}
	}
}
