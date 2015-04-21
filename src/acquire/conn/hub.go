package conn

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan Message

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection

	// gamse
	//games map[int]map[*connection]bool
}

func newHub() *hub {
	return &hub{
		broadcast:   make(chan Message),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
		//games:       make(map[int]map[*connection]bool),
	}
}

func (h *hub) run() {
	var players
	for {
		select {
		case c := <-h.register:
			if players == nil {
				players = make(map[*connection]bool, 6)
			}
			players[c] = true
			n := len(players)
			if n == 6 {
				game.Play(players)
				players = nil
			}
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				// TODO: broadcast message to all connections? here or in conn.go?
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					// Assume client is dead or stuck
					// try defer func() { c.h.unregister <- c }() instead
					delete(h.connections, c)
					close(c.send)
				}
			}
		}
	}
}
