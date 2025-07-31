package jrpc2

import (
	"sync"

	"github.com/creachadair/jrpc2/handler"
	"github.com/creachadair/jrpc2/server"

	"github.com/ecpartan/soap-server-tr069/pkg/jrpc2/methods"
)

type Jrpc2Server struct {
	Server server.Local
}

var Instance *Jrpc2Server

func NewJrpc2Server() *Jrpc2Server {
	once := &sync.Once{}
	once.Do(func() {
		assigner := handler.Map{
			methods.MethodAddScript: handler.New(methods.AddScriptTask),
			methods.MethodGetTree:   handler.New(methods.Get),
		}

		Instance = &Jrpc2Server{
			Server: server.NewLocal(assigner, nil),
		}
	})

	return Instance
}
