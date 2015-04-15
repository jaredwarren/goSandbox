package conn

// NewRouter returns a new router instance.
func NewStockRouter() *SocketRouter {
	return &SocketRouter{}
}

type SocketRouter struct {
	routes []*SocketRoute
}

// NewRoute registers an empty route.
func (r *SocketRouter) NewStockRoute() *SocketRoute {
	route := &SocketRoute{}
	r.routes = append(r.routes, route)
	return route
}

// HandleFunc registers a new route with a matcher for the URL path.
// See Route.Path() and Route.HandlerFunc().
func (r *SocketRouter) HandleFunc(action string, f func(msg)) *SocketRoute {
	return r.NewStockRoute().Action(action).HandlerFunc(f)
}

func (r *SocketRouter) Match(message msg) bool {
	for _, route := range r.routes {
		if route.Match(message) {
			return true
		}
	}
	return false
}
