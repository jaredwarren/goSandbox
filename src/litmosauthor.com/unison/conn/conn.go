package conn

import (
	"fmt"
	//_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/websocket"
	"litmosauthor.com/unison/user"
	"net/http"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// The hub.
	h *hub
}

func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		fmt.Println("reader:", message)
		c.h.broadcast <- message
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
		fmt.Println("writer:", message)
	}
	c.ws.Close()
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type wsHandler struct {
	h *hub
}

type msg struct {
	Num int
}

// TODO: make this a normal funciton like the others, so I can add pass in the user data...? or how do I get the users
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
	c := &connection{send: make(chan []byte, 256), ws: ws, h: wsh.h}
	c.h.register <- c
	defer func() { c.h.unregister <- c }()
	go c.writer()

	m := msg{}

	// broadcast new user
	c.h.broadcast <- []byte("msg|New User:" + userName.Username)

	// blocking
	c.reader()
}
