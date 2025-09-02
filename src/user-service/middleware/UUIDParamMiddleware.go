package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UUIDParamMiddleware(param string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param(param)
		parsed, err := uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Invalid id")
			c.Abort()
			return
		}
		c.Set(param, parsed)
		c.Next()
	}
}
