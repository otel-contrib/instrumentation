package gin

import "github.com/gin-gonic/gin"

// Logger instances a Logger middleware that will write the logs to gin.DefaultWriter.
// By default gin.DefaultWriter = os.Stdout.
func Logger() HandlerFunc {
	return gin.Logger()
}
