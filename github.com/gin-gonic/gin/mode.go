package gin

import "github.com/gin-gonic/gin"

const (
	// DebugMode indicates gin mode is debug.
	DebugMode = gin.DebugMode
	// ReleaseMode indicates gin mode is release.
	ReleaseMode = gin.ReleaseMode
	// TestMode indicates gin mode is test.
	TestMode = gin.TestMode
)

// SetMode sets gin mode according to input string.
func SetMode(value string) {
	gin.SetMode(value)
}
