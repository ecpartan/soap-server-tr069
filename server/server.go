package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ecpartan/soap-server-tr069/db"
	dbb "github.com/ecpartan/soap-server-tr069/db"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/server/handlers/devsoap"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"

	"github.com/ecpartan/soap-server-tr069/internal/config"
	"github.com/ecpartan/soap-server-tr069/internal/devmap"
	logger "github.com/ecpartan/soap-server-tr069/log"
)

// Server a SOAP server, which can be run standalone or used as a http.HandlerFunc
type Server struct {
	mapResponse *devmap.DevMap
	router      *httprouter.Router
	cfg         *config.Config
	db          *db.Service
	httpServer  *http.Server
	cache       *repository.Cache
}

func (s *Server) Register() {
	logger.LogDebug("Registering handlers")

	mainHandler := devsoap.NewHandler(s.mapResponse, s.cache)
	mainHandler.Register(s.router)

	taskHandler := devsoap.NewHandlerCR(s.cache)
	taskHandler.Register(s.router)

	treeHandler := devsoap.NewHandlerGetTree(s.cache)
	treeHandler.Register(s.router)

	userHandler := devsoap.NewHandlerGetUsers(s.db)
	userHandler.Register(s.router)
}

// NewServer construct a new SOAP server
func NewServer(ctx context.Context, cfg *config.Config) (*Server, error) {
	logger.LogDebug("Creating new server")
	router := httprouter.New()

	d, err := dbb.New(ctx, cfg)
	logger.LogDebug("Creating new server", err)

	if err != nil {
		return nil, err
	}

	return &Server{
		mapResponse: devmap.NewDevMap(),
		router:      router,
		cfg:         cfg,
		db:          d,
		cache:       repository.NewCache(ctx, cfg),
	}, nil
}

func (s *Server) RunHTTPServer(ctx context.Context) error {
	logger.LogDebug("Start HTTP server on ip ", s.cfg.Server.Host, ":", s.cfg.Server.Port)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port))
	if err != nil {
		return err
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   s.cfg.Server.CORS.AllowedOrigins,
		AllowedMethods:   s.cfg.Server.CORS.AllowedMethods,
		AllowedHeaders:   s.cfg.Server.CORS.AllowedHeaders,
		AllowCredentials: s.cfg.Server.CORS.AllowCredentials,
	})

	handler := c.Handler(s.router)

	s.httpServer = &http.Server{
		Handler:      handler,
		WriteTimeout: time.Duration(s.cfg.Server.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(s.cfg.Server.ReadTimeout) * time.Second,
	}

	if err = s.httpServer.Serve(listener); err != nil {
		logger.LogDebug("HTTP server stopped")
		return err
	}

	err = s.httpServer.Shutdown(context.Background())
	if err != nil {
		return err
	}

	return err

}
func (s *Server) Run(ctx context.Context) error {

	grp, ctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		return s.RunHTTPServer(ctx)
	})

	return grp.Wait()

}
