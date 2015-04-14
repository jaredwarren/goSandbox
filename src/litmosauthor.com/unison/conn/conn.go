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
	send chan msg

	// The hub.
	h *hub
}

func (c *connection) reader() {
	for {
		message := msg{}
		err := c.ws.ReadJSON(&message)
		if err != nil {
			fmt.Println("error", err)
			// TODO: write filed message?
			break
		}
		fmt.Println("reader:", message)
		c.h.broadcast <- message
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		//err := c.ws.WriteMessage(websocket.TextMessage, message)
		err := c.ws.WriteJSON(message)
		if err != nil {
			// TODO: write filed message?
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
	Action  string
	Message string
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
	c := &connection{send: make(chan msg, 256), ws: ws, h: wsh.h}
	c.h.register <- c
	defer func() { c.h.unregister <- c }()
	go c.writer()

	// broadcast new user
	m := msg{}
	m.Message = "New User:" + userName.Username
	m.Action = "message"
	c.h.broadcast <- m

	// blocking
	c.reader()
}
