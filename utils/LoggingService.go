package utils

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"piped-playfeed/settings"
	"sync"
)

var instance *LoggingService
var mutex sync.Mutex

type LoggingService struct {
	logger               *zap.Logger
	properlyTerminateApp func()
}

func GetLoggingService() *LoggingService {
	if instance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if instance == nil {
			instance = &LoggingService{}
		}
	}
	return instance
}

func (loggingService *LoggingService) InitializeLogger(logFilePath string, debug bool, properlyTerminateApp func()) {
	// remove date time from the builtin logger which is used in addition to Zap
	log.SetFlags(0)

	// init zap
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	writer := zapcore.AddSync(logFile)
	var defaultLogFileLevel zapcore.Level
	if debug {
		defaultLogFileLevel = zapcore.DebugLevel
	} else {
		defaultLogFileLevel = zapcore.WarnLevel
	}
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogFileLevel),
	)
	loggingService.logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// memorize the finalization function
	loggingService.properlyTerminateApp = properlyTerminateApp
}

func (loggingService *LoggingService) Console(msg string) {
	fmt.Println(msg)
}

func (loggingService *LoggingService) ConsoleWarn(msg string) {
	fmt.Println("WARNING: " + msg)
}

func (loggingService *LoggingService) ConsoleFatal(msg string) {
	fmt.Println("ERROR: " + msg)
}

func (loggingService *LoggingService) ConsoleProgress(msg string) {
	if !settings.GetSettingsService().SilentMode {
		loggingService.Console(msg)
	}
}

func (loggingService *LoggingService) Fatal(msg string) {
	loggingService.ConsoleFatal(msg)
	loggingService.properlyTerminateApp()
	loggingService.logger.Fatal(msg)
}

func (loggingService *LoggingService) FatalFromError(err error) {
	loggingService.Fatal(err.Error())
}

func (loggingService *LoggingService) Warn(msg string) {
	loggingService.logger.Warn(msg)
}

func (loggingService *LoggingService) WarnFromError(err error) {
	loggingService.logger.Warn(err.Error())
}

func (loggingService *LoggingService) Info(msg string) {
	loggingService.logger.Info(msg)
}

func (loggingService *LoggingService) Debug(msg string) {
	loggingService.logger.Debug(msg)
}

func (loggingService *LoggingService) SyncLogger() error {
	err := loggingService.logger.Sync()
	if err != nil {
		return err
	}
	return nil
}
