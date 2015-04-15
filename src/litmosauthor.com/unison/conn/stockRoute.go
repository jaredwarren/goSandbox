package conn

// Route stores information to match a request and build URLs.
type SocketRoute struct {
	// Parent where the route was registered (a Router).
	//parent parentRoute
	// Request handler for the route.
	//handler http.Handler
	// List of matchers.
	//matchers []matcher
	// Manager for the variables from host and path.
	//regexp *routeRegexpGroup
	// If true, when the path pattern is "/path/", accessing "/path" will
	// redirect to the former and vice versa.
	//strictSlash bool
	// If true, this route never matches: it is only used to build URLs.
	//buildOnly bool
	// The name used to build URLs.
	//name string
	// Error resulted from building a route.
	//err error
	//

	//buildVarsFunc BuildVarsFunc
	action string

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
	r.handler(message)
	return true
}
