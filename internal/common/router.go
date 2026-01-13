package common

import "github.com/gin-gonic/gin"

var routes = make([]func(string, *gin.Engine), 0)

// RegisterRoute registers a route initialization function
func RegisterRoutes(initFunc func(string, *gin.Engine)) {
	routes = append(routes, initFunc)
}

// InitRouter initializes all registered routes
func InitRouter(base string, r *gin.Engine) {
	for _, routeFunc := range routes {
		routeFunc(base, r)
	}
}
