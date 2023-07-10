package logger

import (
	"updater/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zap.SugaredLogger
}

func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{l.SugaredLogger.With(args...)}
}

func NewLogger(config config.LogConfig) *Logger {
	writeSyncer := getLogWriter(config)
	encoder := getEncoder(config.Format)

	var logLevel zapcore.Level
	err := logLevel.Set(config.Level)
	if err != nil {
		logLevel = zap.InfoLevel
	}

	core := zapcore.NewCore(encoder, writeSyncer, logLevel)

	logger := zap.New(core, zap.AddCaller())

	return &Logger{logger.Sugar()}
}

func getEncoder(format string) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	switch format {
	case "json":
		return zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		return zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return zapcore.NewJSONEncoder(encoderConfig)
	}
}

func getLogWriter(config config.LogConfig) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize, // megabytes
		MaxBackups: 5,
		MaxAge:     config.MaxAge,   //days
		Compress:   config.Compress, // disabled by default
		LocalTime:  true,
	}

	return zapcore.AddSync(lumberJackLogger)
}

// Printf 实现了 gorm.io/gorm/logger.Writer 接口的方法
func (l *Logger) Printf(format string, args ...interface{}) {
	l.Infof(format, args...)
}

// Write 实现了 io.Writer 接口的方法
func (l *Logger) Write(p []byte) (n int, err error) {
	l.Info(string(p))
	return len(p), nil
}

func (l *Logger) Println(args ...interface{}) {
	l.Info(args...)
}

var defaultLogger *Logger

func InitLogger() {
	defaultLogger = NewLogger(config.GetConfig().LogConfig)
}

func GetLogger() *Logger {
	return defaultLogger
}

func Printf(format string, args ...interface{}) {
	defaultLogger.Printf(format, args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

func Println(args ...interface{}) {
	defaultLogger.Info(args...)
}
