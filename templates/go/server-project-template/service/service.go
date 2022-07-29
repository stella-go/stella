package service

type HelloService struct {
}

func (p *HelloService) Hello() string {
	return "Hello!"
}
