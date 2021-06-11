package rotater

import "github.com/gin-gonic/gin"

type HelloRotate struct{}

func (p *HelloRotate) Rotater() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"GET /hello": p.Hello,
	}
}

func (p *HelloRotate) Hello(c *gin.Context) {
	c.String(200, "Hello.")
}
