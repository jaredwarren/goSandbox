package game

import (
//"fmt"
//"github.com/gorilla/mux"
//"net/http"
//"acquire/conn"
)

type Player struct {
	// /Conn *conn.Connection
	Name string
}

/*// Check wethever the player is still connected by sending a ping command.
func (p *Player) Alive() bool {
	if err := p.Conn.ws.WriteJSON(message)(Message{Action: "ping"}); err != nil {
		return false
	}
	message := Message{}
	if err := p.Conn.ws.ReadJSON(&message); err != nil {
		return false
	}
	return message.Action == "pong"
}

func (p *Player) Send(msg Message) {
	fmt.Println("...")
	if p.Conn != nil {
		p.send <- msg
	}
}*/
