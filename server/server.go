package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	_ "github.com/ecpartan/soap-server-tr069/docs"
	"github.com/ecpartan/soap-server-tr069/pkg/jrpc2"
	"github.com/ecpartan/soap-server-tr069/pkg/jrpc2/middleware"
	usecase_service "github.com/ecpartan/soap-server-tr069/repository/db/domain/usecase/service"
	"github.com/ecpartan/soap-server-tr069/repository/storage"
	"github.com/ecpartan/soap-server-tr069/tasks/tasker"
	"github.com/ecpartan/soap-server-tr069/web"

	"github.com/ecpartan/soap-server-tr069/pkg/monitoring"
	"github.com/ecpartan/soap-server-tr069/pkg/users/login"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/soap-server-tr069/server/handlers/devsoap"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	swag "github.com/swaggo/http-swagger"
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
	service     *usecase_service.Service
	httpServer  *http.Server
	cache       *repository.Cache
	jrpc2Server *jrpc2.Jrpc2Server
	ExecTasks   *tasker.Tasker
	monitoring  *monitoring.MetricsService
}

func (s *Server) Register() {
	logger.LogDebug("Registering handlers")

	mainHandler := devsoap.NewHandler(s.mapResponse, s.cache, s.service.DeviceService, s.ExecTasks, s.monitoring)
	mainHandler.Register(s.router)

	taskHandler := devsoap.NewHandlerCR(s.cache, s.ExecTasks, s.service.TasksService, s.service.DeviceService)
	taskHandler.Register(s.router)

	treeHandler := devsoap.NewHandlerGetTree(s.cache)
	treeHandler.Register(s.router)

	userHandler := devsoap.NewHandlerGetUsers(s.service.UserService)
	userHandler.Register(s.router)

	loginHandler := login.NewHandler(s.service.UserService)
	loginHandler.Register(s.router)

	frontHandler := middleware.NewHandler(s.cache, s.ExecTasks)
	frontHandler.Register(s.router)
}

var once sync.Once
var Instance *Server

// NewServer construct a new SOAP server
func NewServer(ctx context.Context, cfg *config.Config) (*Server, error) {
	logger.LogDebug("Creating new server")
	router := httprouter.New()

	logger.LogDebug("swagger docs initializing")
	router.Handler(http.MethodGet, "/swagger", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))
	router.Handler(http.MethodGet, "/swagger/*any", swag.WrapHandler)
	cache := repository.NewCache(ctx, cfg)

	mservice, err := monitoring.NewMetricsService(router)
	if err != nil {
		logger.LogDebug("Monitoring service not initialized", err)
		return nil, err
	}

	web.Register(cfg.Server.Host, cfg.Server.PortWeb)
	dbstorage, err := storage.NewStorage(&cfg.DatabaseConf)

	if err != nil {
		return nil, err
	}

	serviceDevice := usecase_service.NewService(dbstorage)

	var once sync.Once

	once.Do(func() {
		Instance = &Server{
			mapResponse: devmap.NewDevMap(),
			router:      router,
			cfg:         cfg,
			service:     serviceDevice,
			cache:       cache,
			jrpc2Server: jrpc2.NewJrpc2Server(),
			ExecTasks:   tasker.InitTasker(dbstorage),
			monitoring:  mservice,
		}
	})
	logger.LogDebug("GetServer", Instance)

	return Instance, nil
}

func (s *Server) RunHTTPServer(ctx context.Context) error {
	logger.LogDebug("Start HTTP server on ip ", s.cfg.Server.Host, ":", s.cfg.Server.Port)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port))
	logger.LogDebug("qqqq", err)

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

	err = s.httpServer.Shutdown(ctx)
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
