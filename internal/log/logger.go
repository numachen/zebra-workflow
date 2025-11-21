package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	mu     sync.Mutex
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger

	// keep current config for introspection
	currentLevel    string
	currentEncoding string
	currentOutputs  []string
)

// Init 初始化全局 zap logger。
// level: "debug","info","warn","error"
// encoding: "json" or "console"
// outputs: []string, e.g. {"stdout","logs/app.log"}
func Init(level string, encoding string, outputs []string) error {
	mu.Lock()
	defer mu.Unlock()

	cfg, err := buildZapConfig(level, encoding, outputs)
	if err != nil {
		return err
	}
	l, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("build zap logger: %w", err)
	}

	// swap in new logger
	if Logger != nil {
		_ = Logger.Sync()
	}
	Logger = l
	Sugar = l.Sugar()

	currentLevel = level
	currentEncoding = encoding
	currentOutputs = append([]string(nil), outputs...)

	return nil
}

// Reload 用于热重载日志配置（安全替换当前 logger）
func Reload(level string, encoding string, outputs []string) error {
	mu.Lock()
	defer mu.Unlock()

	cfg, err := buildZapConfig(level, encoding, outputs)
	if err != nil {
		return err
	}
	l, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("build zap logger during reload: %w", err)
	}

	// swap
	old := Logger
	Logger = l
	Sugar = l.Sugar()

	// update current config
	currentLevel = level
	currentEncoding = encoding
	currentOutputs = append([]string(nil), outputs...)

	// flush and close old logger
	if old != nil {
		_ = old.Sync()
	}

	Sugar.Infow("logger reloaded", "level", level, "encoding", encoding, "outputs", outputs)
	return nil
}

// Sync flushes any buffered log entries.
func Sync() {
	mu.Lock()
	defer mu.Unlock()
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// Close is alias to Sync for convenience.
func Close() {
	Sync()
}

// buildZapConfig 构造 zap.Config
func buildZapConfig(level string, encoding string, outputs []string) (zap.Config, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		// default to info if invalid
		zapLevel = zapcore.InfoLevel
		level = zapLevel.String()
	}

	enc := strings.ToLower(encoding)
	if enc != "json" && enc != "console" {
		enc = "json"
	}

	// prepare outputs: map "stdout"/"stderr" to themselves; for files ensure directory exists
	outPaths := make([]string, 0, len(outputs))
	for _, o := range outputs {
		if o == "stdout" || o == "stderr" {
			outPaths = append(outPaths, o)
			continue
		}
		// create dir if needed
		dir := filepath.Dir(o)
		if dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				// if create dir fails, fallback to stdout and continue
				outPaths = append(outPaths, "stdout")
				continue
			}
		}
		outPaths = append(outPaths, o)
	}
	if len(outPaths) == 0 {
		outPaths = []string{"stdout"}
	}

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapLevel),
		Development: false,
		Encoding:    enc,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      outPaths,
		ErrorOutputPaths: outPaths,
	}
	return cfg, nil
}
