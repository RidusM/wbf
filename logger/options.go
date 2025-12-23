package logger

import (
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type GlobalConfig struct {
	Level      Level
	AppName    string
	Env        string
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	Stdout     bool
}

type Option func(*GlobalConfig)

func defaultConfigs() *GlobalConfig {
	return &GlobalConfig{
		Level:      InfoLevel,
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     30,
		Compress:   true,
		Stdout:     true,
	}
}

func WithLevel(l Level) Option { return func(c *GlobalConfig) { c.Level = l } }

func WithRotation(filename string, maxSize, maxBackups, maxAge int) Option {
	return func(c *GlobalConfig) {
		c.Filename = filename
		c.MaxSize = maxSize
		c.MaxBackups = maxBackups
		c.MaxAge = maxAge
	}
}

func (c *GlobalConfig) GetWriter() io.Writer {
	var writers []io.Writer
	if c.Stdout {
		writers = append(writers, os.Stdout)
	}
	if c.Filename != "" {
		writers = append(writers, &lumberjack.Logger{
			Filename:   c.Filename,
			MaxSize:    c.MaxSize,
			MaxBackups: c.MaxBackups,
			MaxAge:     c.MaxAge,
			Compress:   c.Compress,
		})
	}
	return io.MultiWriter(writers...)
}
