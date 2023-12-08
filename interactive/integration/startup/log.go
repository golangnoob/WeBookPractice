package startup

import "webooktrial/pkg/logger"

func InitLog() logger.LoggerV1 {
	return &logger.NopLogger{}
}
