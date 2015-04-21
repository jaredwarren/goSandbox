package conn

import (
	"fmt"
	//_ "github.com/go-sql-driver/mysql"

	"acquire/game"
	"github.com/gorilla/websocket"
	"litmosauthor.com/unison/user"
	"net/http"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan Message

	// The hub.
	h  *hub
	wr *SocketRouter
}

func (c *connection) reader() {
	for {
		message := Message{}
		err := c.ws.ReadJSON(&message)
		if err != nil {
			fmt.Println("ReadJSON Error", err)
			// TODO: write filed message?
			break
		}

		//TODO: see if I can create some sort of router to route actions...
		if !c.wr.Match(message) {
			fmt.Println("No action:", message)
		} else {
			fmt.Println("<=", message)
			c.h.broadcast <- message
		}
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		//err := c.ws.WriteMessage(websocket.TextMessage, message)
		err := c.ws.WriteJSON(message)
		if err != nil {
			fmt.Println("writerError:", err)
			// TODO: write filed message?
			break
		}
		fmt.Println("=>", message)
	}
	c.ws.Close()
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type wsHandler struct {
	h *hub
}

func (wsh wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// make sure session is valid
	userName := user.GetUserName(r)
	if userName == nil {
		fmt.Println("No User")
		// TODO: throw error
		return
	}
	// setup websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	// TODO: make player instead of connection
	//c := &connection{send: make(chan Message, 256), ws: ws, h: wsh.h, wr: wsh.wr}
	player := &Player{}
	c.h.register <- c
	defer func() { c.h.unregister <- c }()
	go c.writer()

	// broadcast new user
	m := Message{Text: "New User:" + userName.Username, Action: "message"}
	c.h.broadcast <- m

	// blocking
	c.reader()
}
