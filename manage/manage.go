package manage

import "github.com/chuccp/d-mail/core"

type manage interface {
	Init(context *core.Context, server core.IHttpServer)
}
type Server struct {
	manages []manage
	core.IHttpServer
	context *core.Context
}

func NewServer() *Server {
	server := &Server{}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	return server
}

func (s *Server) Name() string {
	return "manage"
}
func (s *Server) addManage(manage manage) {
	s.manages = append(s.manages, manage)
}
func (s *Server) Init(context *core.Context) {
	s.context = context
	s.manages = make([]manage, 0)
	s.addManage(&Stmp{})
	s.addManage(&Mail{})
	s.addManage(&Set{})
	s.addManage(&Token{})
	for _, a := range s.manages {
		a.Init(s.context, s.IHttpServer)
	}
}
