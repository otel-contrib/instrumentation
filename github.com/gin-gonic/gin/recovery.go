package gin

import "github.com/gin-gonic/gin"

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() HandlerFunc {
	return gin.Recovery()
}
