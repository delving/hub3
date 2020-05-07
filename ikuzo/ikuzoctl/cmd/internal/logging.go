package internal

import "github.com/delving/hub3/ikuzo/logger"

type Logging struct {
	DevMode       bool   `json:"devmode"`
	Level         string `json:"level"`
	WithCaller    bool   `json:"withCaller"`
	ConsoleLogger bool   `json:"consoleLogger"`
}

func (l *Logging) GetConfig() logger.Config {
	return logger.Config{
		LogLevel:            logger.ParseLogLevel(l.Level),
		WithCaller:          l.WithCaller,
		EnableConsoleLogger: l.ConsoleLogger,
	}
}
