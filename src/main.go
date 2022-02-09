package main

import (
	"github.com/wonderivan/logger"
)

func main() {
	for {
		sendHello()
	}
}
func checkError(err error) {
	if err != nil {
		logger.Error(err)
	}
}

// Is used if the return is ignored but the error is not
func checkErrorReturn(_ interface{}, err error) {
	if err != nil {
		logger.Error(err)
	}
}
