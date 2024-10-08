package core

import (
	"github.com/chuccp/smtp2http/config"
	"github.com/chuccp/smtp2http/login"
	"github.com/chuccp/smtp2http/web"
	"go.uber.org/zap"
	"log"
)

type DMail struct {
	context    *Context
	httpServer *web.HttpServer
	log        *zap.Logger
	config     *config.Config
	servers    []Server
	webPort    int
	apiPort    int
}

func Create() *DMail {
	return &DMail{webPort: 0, apiPort: 0, servers: make([]Server, 0), config: config.NewConfig()}
}
func (m *DMail) AddServer(server Server) {
	m.servers = append(m.servers, server)
}

func (m *DMail) startHttpServer() error {
	port := m.context.GetCfgInt("manage", "port")
	certFile := m.context.GetCfgString("manage", "certFile")
	keyFile := m.context.GetCfgString("manage", "keyFile")
	m.context.log.Info("startHttpServer", zap.String("name", "manage"), zap.Int("port", port))
	err := m.httpServer.StartAutoTLS(port, certFile, keyFile)
	if err != nil {
		m.context.log.Error("服务启动失败", zap.String("name", "DMail"), zap.Int("port", port), zap.Error(err))
		return err
	}
	return nil
}
func (m *DMail) Start(webPort int, apiPort int) {
	m.webPort = webPort
	m.apiPort = apiPort
	m.reStart()
}
func (m *DMail) reStart() {
	err := m.config.Init(m.webPort, m.apiPort)
	if err != nil {
		log.Panic(err)
		return
	}
	logPath := m.config.GetStringOrDefault("log", "filename", "run.log")
	logger, err := initLogger(logPath)
	if err != nil {
		log.Panic(err)
		return
	}
	m.log = logger
	m.context = &Context{log: m.log, config: m.config}
	digestAuth := login.NewDigestAuth(m.context.SecretProvider)
	m.context.digestAuth = digestAuth
	m.httpServer = web.NewServer(digestAuth)
	m.context.httpServer = m.httpServer
	isInit := m.config.GetBooleanOrDefault("core", "init", false)
	if isInit {
		err := m.context.initDb()
		if err != nil {
			m.log.Panic("initDb", zap.Error(err))
			return
		}
	}
	for _, server := range m.servers {
		if s, ok := server.(IHttpServer); ok {
			s.init(m.context)
		}
		server.Init(m.context)
		if s, ok := server.(IHttpServer); ok {
			if !s.useCorePort() {
				go func() {
					err := s.start()
					if err != nil {
						log.Panic(err)
					}
				}()
			}
		}
	}
	err = m.startHttpServer()
	if err != nil {
		m.log.Panic("Start", zap.Error(err))
		return
	}
}
