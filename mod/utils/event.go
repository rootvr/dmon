package utils

import (
	logger "dmon/mod/logger"
)

func Panic(module string, err error, exec string, messages ...string) {
	if err != nil {
		logger.Error(module, exec, messages)
		panic(err)
	}
}
