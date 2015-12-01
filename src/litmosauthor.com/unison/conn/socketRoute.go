package conn

// Route stores information to match a request and build URLs.
type SocketRoute struct {
	action  string
	handler func(msg)
}

func (r *SocketRoute) Action(tpl string) *SocketRoute {
	r.action = tpl
	return r
}

// HandlerFunc sets a handler function for the route.
func (r *SocketRoute) HandlerFunc(f func(msg)) *SocketRoute {
	r.handler = f
	return r
}

// HandlerFunc sets a handler function for the route.
func (r *SocketRoute) Match(message msg) bool {
	if r.action == message.Action {
		r.handler(message)
		return true
	} else {
		return false
	}
}
