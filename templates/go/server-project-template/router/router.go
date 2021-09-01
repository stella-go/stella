package router

import "github.com/gin-gonic/gin"

type HelloRouter struct{}

func (p *HelloRouter) Router() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"GET /hello": p.Hello,
	}
}

func (p *HelloRouter) Hello(c *gin.Context) {
	c.String(200, "Hello.")
}
