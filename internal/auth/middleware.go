package auth

import (
	// "net/http" // Removed unused import

	"github.com/gin-gonic/gin"
)

// BasicAuth returns a Gin middleware function for Basic Authentication.
// It checks credentials against the provided username and password.
func BasicAuth(username, password string) gin.HandlerFunc {
	// Ensure credentials are provided during setup
	if username == "" || password == "" {
		panic("Username and password for basic auth cannot be empty")
	}

	expectedCredentials := gin.Accounts{
		username: password,
	}

	return gin.BasicAuth(expectedCredentials)
}

// Optional: A simpler custom implementation if gin.BasicAuth is not flexible enough
// func CustomBasicAuth(expectedUser, expectedPass string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		user, pass, hasAuth := c.Request.BasicAuth()
//
// 		if !hasAuth || user != expectedUser || pass != expectedPass {
// 			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 			return
// 		}
//
// 		c.Next()
// 	}
// }
