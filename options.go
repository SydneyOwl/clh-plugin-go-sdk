package clhplugin

import (
	"errors"
	"runtime"

	"time"
)

const (
	defaultPipePathUnix    = "/tmp/clh.plugin"
	defaultPipePathWindows = "\\\\.\\pipe\\clh.plugin"
	defaultHeartbeat       = 5 * time.Second
	defaultRequestTimeout  = 8 * time.Second
	defaultWaitBuffer      = 256
	defaultSDKName         = "clh-plugin-go-sdk"
	defaultSDKVersion      = "v20260312"
)

type MessageHandler func(Message)

type Option func(*Config) error

type Config struct {
	PipePath          string
	HeartbeatInterval time.Duration
	RequestTimeout    time.Duration
	WaitBufferSize    int
	OnMessage         MessageHandler
}

func defaultConfig() Config {
	pipePath := defaultPipePathUnix
	if runtime.GOOS == "windows" {
		pipePath = defaultPipePathWindows
	}
	return Config{
		PipePath:          pipePath,
		HeartbeatInterval: defaultHeartbeat,
		RequestTimeout:    defaultRequestTimeout,
		WaitBufferSize:    defaultWaitBuffer,
	}
}

func WithPipePath(pipePath string) Option {
	return func(cfg *Config) error {
		if pipePath == "" {
			return errors.New("pipe path cannot be empty")
		}
		cfg.PipePath = pipePath
		return nil
	}
}

func WithHeartbeatInterval(interval time.Duration) Option {
	return func(cfg *Config) error {
		if interval < 0 {
			return errors.New("heartbeat interval cannot be negative")
		}
		cfg.HeartbeatInterval = interval
		return nil
	}
}

func WithMessageHandler(handler MessageHandler) Option {
	return func(cfg *Config) error {
		cfg.OnMessage = handler
		return nil
	}
}

func WithRequestTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout <= 0 {
			return errors.New("request timeout must be positive")
		}
		cfg.RequestTimeout = timeout
		return nil
	}
}

func WithWaitBufferSize(size int) Option {
	return func(cfg *Config) error {
		if size <= 0 {
			return errors.New("wait buffer size must be greater than 0")
		}
		cfg.WaitBufferSize = size
		return nil
	}
}
