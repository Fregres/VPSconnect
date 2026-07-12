package httpapi

import "net/http"

type Server struct {
	token string
}

func NewServer(token string) *Server {
	return &Server{token: token}
}

func (s *Server) Handler() http.Handler {
	return s.routes()
}
