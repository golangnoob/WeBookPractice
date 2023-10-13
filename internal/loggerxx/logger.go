package loggerxx

import (
	"go.uber.org/zap"
)

//  没有必要自己维护一个包变量，多此一举，和直接使用zap.L()无甚差别

var Logger *zap.Logger

func InitLogger(l *zap.Logger) {
	Logger = l
}

// InitLoggerV1 main 函数调用一下
func InitLoggerV1() {
	Logger, _ = zap.NewDevelopment()
}

//var SecureLogger *zap.Logger
