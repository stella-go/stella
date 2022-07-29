package router

import (
	"{{ project-name }}/service"
	"github.com/gin-gonic/gin"
)

type HelloRouter struct {
	HelloService *service.HelloService
}

func (p *HelloRouter) Router() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"GET /hello": p.Hello,
	}
}

func (p *HelloRouter) Hello(c *gin.Context) {
	s := p.HelloService.Hello()
	c.String(200, s)
}
