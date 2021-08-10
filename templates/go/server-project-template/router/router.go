package router

import "github.com/gin-gonic/gin"

type HelloRoute struct{}

func (p *HelloRoute) Router() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"GET /hello": p.Hello,
	}
}

func (p *HelloRoute) Hello(c *gin.Context) {
	c.String(200, "Hello.")
}
