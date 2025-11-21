package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"zebra-workflow/internal/handler"
	logger "zebra-workflow/internal/log"
	"zebra-workflow/internal/temporal"
)

// Config 映射 configs/config.yaml 中需要的部分
type Config struct {
	App struct {
		Name string `yaml:"name" json:"name"`
	} `yaml:"app" json:"app"`

	HTTP struct {
		Addr string `yaml:"addr" json:"addr"`
	} `yaml:"http" json:"http"`

	Logging struct {
		Level    string   `yaml:"level" json:"level"`
		Encoding string   `yaml:"encoding" json:"encoding"`
		Outputs  []string `yaml:"outputs" json:"outputs"`
	} `yaml:"logging" json:"logging"`
}

func main() {
	const cfgPath = "configs/config.yaml"

	// 1. load config
	var cfg Config
	if err := conf.Load(cfgPath, &cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. init logger from config (level, encoding, outputs)
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.Encoding, cfg.Logging.Outputs); err != nil {
		// fallback to sensible defaults
		_ = logger.Init("info", "json", []string{"stdout"})
	}
	defer logger.Close()
	logger.Sugar.Infof("starting %s", cfg.App.Name)

	// 3. parse addr
	addr := cfg.HTTP.Addr
	if strings.TrimSpace(addr) == "" {
		addr = ":8888"
	}
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		if _, perr := strconv.Atoi(addr); perr == nil {
			host = "0.0.0.0"
			portStr = addr
		} else {
			logger.Sugar.Warnf("invalid http.addr %q, falling back to :8888", addr)
			host = "0.0.0.0"
			portStr = "8888"
		}
	}
	if host == "" {
		host = "0.0.0.0"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Sugar.Warnf("invalid port %q, falling back to 8888", portStr)
		port = 8888
	}

	restCfg := rest.RestConf{
		Host: host,
		Port: port,
	}

	server := rest.MustNewServer(restCfg)
	defer server.Stop()

	// 4. init temporal client (reads configs/config.yaml)
	tc, err := temporal.NewClientFromConfig(cfgPath)
	if err != nil {
		logger.Sugar.Fatalf("temporal client init failed: %v", err)
	}
	defer tc.Close()

	// 5. register handlers (handler uses temporal client)
	handler.RegisterRoutes(server, tc)

	fullAddr := fmt.Sprintf("%s:%d", host, port)
	logger.Sugar.Infof("HTTP server started on %s", fullAddr)

	// 6. start config watcher for hot reload of logging config
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			logger.Sugar.Errorw("failed to create fsnotify watcher", "err", err)
			return
		}
		defer watcher.Close()

		if err := watcher.Add(cfgPath); err != nil {
			logger.Sugar.Errorw("failed to add config file to watcher", "path", cfgPath, "err", err)
			return
		}

		debounce := time.NewTimer(0)
		<-debounce.C // drain
		var pending bool

		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				// watch for write/create/rename
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
					// debounce rapid events
					if !pending {
						debounce.Reset(200 * time.Millisecond)
						pending = true
					}
				}
			case <-debounce.C:
				if pending {
					var newCfg Config
					if err := conf.Load(cfgPath, &newCfg); err != nil {
						logger.Sugar.Errorw("failed to reload config on change", "err", err)
					} else {
						// reload logging config only
						if err := logger.Reload(newCfg.Logging.Level, newCfg.Logging.Encoding, newCfg.Logging.Outputs); err != nil {
							logger.Sugar.Errorw("failed to reload logger", "err", err)
						} else {
							logger.Sugar.Infow("logging configuration reloaded from file", "path", cfgPath)
						}
					}
					pending = false
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Sugar.Errorw("fsnotify watcher error", "err", err)
			}
		}
	}()

	// start server (blocks)
	server.Start()

	<-context.Background().Done()
}
