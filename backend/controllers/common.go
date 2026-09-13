package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RequestUser struct {
	ID          uint
	Username    string
	Role        string
	CompanyName string
}

func respondError(c *gin.Context, status int, err error) {
	if err == nil {
		c.JSON(status, gin.H{"error": http.StatusText(status)})
		return
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

func respondErrorMessage(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func getRequestUser(c *gin.Context) (RequestUser, error) {
	var u RequestUser

	idVal, ok := c.Get("user_id")
	if !ok {
		return u, fmt.Errorf("missing user_id in context")
	}
	switch id := idVal.(type) {
	case uint:
		u.ID = id
	case int:
		if id < 0 {
			return u, fmt.Errorf("invalid user_id value")
		}
		u.ID = uint(id)
	case float64:
		if id < 0 {
			return u, fmt.Errorf("invalid user_id value")
		}
		u.ID = uint(id)
	default:
		return u, fmt.Errorf("invalid user_id type")
	}

	if username, ok := c.Get("username"); ok {
		if s, ok := username.(string); ok {
			u.Username = s
		}
	}
	if role, ok := c.Get("role"); ok {
		if s, ok := role.(string); ok {
			u.Role = s
		}
	}
	if company, ok := c.Get("company_name"); ok {
		if s, ok := company.(string); ok {
			u.CompanyName = s
		}
	}

	return u, nil
}

func parsePositiveIntQuery(c *gin.Context, key string, defaultVal, maxVal int) int {
	raw := c.Query(key)
	if raw == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultVal
	}
	if maxVal > 0 && n > maxVal {
		return maxVal
	}
	return n
}
