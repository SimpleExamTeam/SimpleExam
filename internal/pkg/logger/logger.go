package logger

import (
	"exam-system/internal/config"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var (
	Logger    *log.Logger
	logLevel  LogLevel
	logFormat string
)

// Setup 初始化日志系统
func Setup() error {
	// 获取配置
	cfg := config.GlobalConfig.Log

	// 设置日志级别
	switch strings.ToLower(cfg.Level) {
	case "debug":
		logLevel = DEBUG
	case "info":
		logLevel = INFO
	case "warn":
		logLevel = WARN
	case "error":
		logLevel = ERROR
	case "fatal":
		logLevel = FATAL
	default:
		return fmt.Errorf("invalid log level: %s", cfg.Level)
	}

	// 设置日志格式
	logFormat = strings.ToLower(cfg.Format)
	if logFormat != "json" && logFormat != "text" {
		return fmt.Errorf("invalid log format: %s", cfg.Format)
	}

	// 设置输出方式
	var writer io.Writer
	switch strings.ToLower(cfg.Output) {
	case "console":
		writer = os.Stdout
	case "file":
		fileWriter, err := setupFileWriter(cfg)
		if err != nil {
			return err
		}
		writer = fileWriter
	case "both":
		fileWriter, err := setupFileWriter(cfg)
		if err != nil {
			return err
		}
		writer = io.MultiWriter(os.Stdout, fileWriter)
	default:
		return fmt.Errorf("invalid log output: %s", cfg.Output)
	}

	// 创建logger
	Logger = log.New(writer, "", 0)

	Info("Logger initialized successfully")
	return nil
}

// setupFileWriter 设置文件输出
func setupFileWriter(logCfg struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}) (io.Writer, error) {
	// 确保日志目录存在
	logDir := filepath.Dir(logCfg.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// 打开日志文件
	file, err := os.OpenFile(logCfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return file, nil
}

// formatMessage 格式化日志消息
func formatMessage(level string, msg string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if logFormat == "json" {
		return fmt.Sprintf(`{"time":"%s","level":"%s","msg":"%s"}`, timestamp, level, msg)
	}
	return fmt.Sprintf("[%s] %s: %s", timestamp, level, msg)
}

// shouldLog 检查是否应该记录日志
func shouldLog(level LogLevel) bool {
	return level >= logLevel
}

// GetLogger 获取日志实例
func GetLogger() *log.Logger {
	if Logger == nil {
		// 如果日志未初始化，使用默认配置
		Logger = log.New(os.Stdout, "", 0)
	}
	return Logger
}

// 便捷方法
func Debug(args ...interface{}) {
	if shouldLog(DEBUG) {
		msg := fmt.Sprint(args...)
		GetLogger().Print(formatMessage("DEBUG", msg))
	}
}

func Debugf(format string, args ...interface{}) {
	if shouldLog(DEBUG) {
		msg := fmt.Sprintf(format, args...)
		GetLogger().Print(formatMessage("DEBUG", msg))
	}
}

func Info(args ...interface{}) {
	if shouldLog(INFO) {
		msg := fmt.Sprint(args...)
		GetLogger().Print(formatMessage("INFO", msg))
	}
}

func Infof(format string, args ...interface{}) {
	if shouldLog(INFO) {
		msg := fmt.Sprintf(format, args...)
		GetLogger().Print(formatMessage("INFO", msg))
	}
}

func Warn(args ...interface{}) {
	if shouldLog(WARN) {
		msg := fmt.Sprint(args...)
		GetLogger().Print(formatMessage("WARN", msg))
	}
}

func Warnf(format string, args ...interface{}) {
	if shouldLog(WARN) {
		msg := fmt.Sprintf(format, args...)
		GetLogger().Print(formatMessage("WARN", msg))
	}
}

func Error(args ...interface{}) {
	if shouldLog(ERROR) {
		msg := fmt.Sprint(args...)
		GetLogger().Print(formatMessage("ERROR", msg))
	}
}

func Errorf(format string, args ...interface{}) {
	if shouldLog(ERROR) {
		msg := fmt.Sprintf(format, args...)
		GetLogger().Print(formatMessage("ERROR", msg))
	}
}

func Fatal(args ...interface{}) {
	if shouldLog(FATAL) {
		msg := fmt.Sprint(args...)
		GetLogger().Print(formatMessage("FATAL", msg))
		os.Exit(1)
	}
}

func Fatalf(format string, args ...interface{}) {
	if shouldLog(FATAL) {
		msg := fmt.Sprintf(format, args...)
		GetLogger().Print(formatMessage("FATAL", msg))
		os.Exit(1)
	}
}
