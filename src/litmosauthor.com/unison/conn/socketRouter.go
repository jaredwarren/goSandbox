package conn

// NewRouter returns a new router instance.
func NewStockRouter() *SocketRouter {
	return &SocketRouter{}
}

type SocketRouter struct {
	// Configurable Handler to be used when no route matches.
	//NotFoundHandler http.Handler
	// Parent route, if this is a subrouter.
	//parent parentRoute
	// Routes to be matched, in order.
	routes []*SocketRoute
	// Routes by name for URL building.
	//namedRoutes map[string]*SocketRoute
	// See Router.StrictSlash(). This defines the flag for new routes.
	//strictSlash bool
	// If true, do not clear the request context after handling the request
	//KeepContext bool
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
