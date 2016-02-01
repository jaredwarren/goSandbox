package conn

import (
	"acquire/user"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan msg

	// The hub.
	h  *hub
	wr *SocketRouter

	u *user.User
}

func (c *connection) reader() {
	defer func() {
		c.h.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		message := msg{}
		if err := c.ws.ReadJSON(&message); err != nil {
			fmt.Println("ReadJSON Error", err)
			break
		}

		message.Sender = c.u.Username

		// rebradsast message to others
		if !c.wr.Match(message) {
			fmt.Println("No matching action:", message)
		} else {
			fmt.Println("<=", message)
			c.h.broadcast <- message
		}
	}
}

func (c *connection) writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.writeJson(message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// write writes a message with the given message type and payload.
func (c *connection) writeJson(message msg) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	fmt.Println("=>", message)
	return c.ws.WriteJSON(message)
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type wsHandler struct {
	h  *hub
	wr *SocketRouter
}

type msg struct {
	Action  string
	Message string
	Sender  string
}

func (wsh wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// make sure session is valid
	user := user.GetUserName(r)
	if user == nil {
		fmt.Println("No User")
		// TODO: throw error
		return
	}

	// setup websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &connection{send: make(chan msg, 256), ws: ws, h: wsh.h, wr: wsh.wr, u: user}
	c.h.register <- c
	defer func() { c.h.unregister <- c }()

	// start writer
	go c.writer()

	// broadcast new user
	m := msg{Message: "New User:" + user.Username, Action: "message"}
	c.h.broadcast <- m

	// start reader, blocking
	c.reader()
}
