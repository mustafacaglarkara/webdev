package router

import (
	"github.com/gorilla/mux"
)

// RegisterMuxRoutes walks a gorilla/mux router and registers named routes into the router's template map.
// Routes must be named via router.HandleFunc(...).Name("routename") for this to pick them up.
func RegisterMuxRoutes(r *mux.Router) error {
	return r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		name := route.GetName()
		if name == "" {
			return nil
		}
		path, err := route.GetPathTemplate()
		if err != nil {
			// ignore unnamed or complex routes we cannot get a template for
			return nil
		}
		RegisterRoute(name, path)
		RegisterTemplate(name, path)
		return nil
	})
}
