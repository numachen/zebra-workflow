package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/zeromicro/go-zero/core/conf"
	"go.temporal.io/sdk/worker"

	"zebra-workflow/internal/activity"
	logger "zebra-workflow/internal/log"
	"zebra-workflow/internal/temporal"
	"zebra-workflow/internal/workflow"
)

type WorkerConfig struct {
	Logging struct {
		Level    string   `yaml:"level" json:"level"`
		Encoding string   `yaml:"encoding" json:"encoding"`
		Outputs  []string `yaml:"outputs" json:"outputs"`
	} `yaml:"logging" json:"logging"`
}

func main() {
	const cfgPath = "configs/config.yaml"

	// load config to get logging level
	var cfg WorkerConfig
	if err := conf.Load(cfgPath, &cfg); err != nil {
		// fallback
		_ = logger.Init("info", "json", []string{"stdout"})
	} else {
		if err := logger.Init(cfg.Logging.Level, cfg.Logging.Encoding, cfg.Logging.Outputs); err != nil {
			_ = logger.Init("info", "json", []string{"stdout"})
		}
	}
	defer logger.Close()

	// start watcher for config reload (logging)
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
		<-debounce.C
		var pending bool

		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
					if !pending {
						debounce.Reset(200 * time.Millisecond)
						pending = true
					}
				}
			case <-debounce.C:
				if pending {
					var newCfg WorkerConfig
					if err := conf.Load(cfgPath, &newCfg); err != nil {
						logger.Sugar.Errorw("failed to reload worker config", "err", err)
					} else {
						if err := logger.Reload(newCfg.Logging.Level, newCfg.Logging.Encoding, newCfg.Logging.Outputs); err != nil {
							logger.Sugar.Errorw("failed to reload logger in worker", "err", err)
						} else {
							logger.Sugar.Infow("worker logging configuration reloaded", "path", cfgPath)
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

	// init temporal client
	tc, err := temporal.NewClientFromConfig(cfgPath)
	if err != nil {
		logger.Sugar.Fatalf("temporal client init failed: %v", err)
	}
	defer tc.Close()

	// create worker using exported client
	w := worker.New(tc.Client(), tc.DefaultQueue(), worker.Options{})

	// register workflows and activities into the worker
	for _, wf := range workflow.ListRegistered() {
		workflowFunc := wf.Factory()
		logger.Sugar.Infof("registering workflow name=%s version=%s", wf.Name, wf.Version)
		w.RegisterWorkflowWithOptions(workflowFunc, workflow.GetRegisterOptions(wf))
	}

	act := activity.NewActivityImpl()
	w.RegisterActivity(act)
	// 注册 DSL 示例活动（函数包装），使活动类型名为 "SampleActivity" 等，匹配 DSL YAML 中的 a.Name
	w.RegisterActivity(activity.SampleActivity)
	w.RegisterActivity(activity.SampleActivitySendEmail)
	w.RegisterActivity(activity.SampleActivitySendEmailTyped)
	w.RegisterActivity(activity.GetTitle)

	// start worker
	if err := w.Start(); err != nil {
		logger.Sugar.Fatalf("worker start failed: %v", err)
	}
	fmt.Println("Temporal worker started")
	select {
	case <-context.Background().Done():
	}
}
