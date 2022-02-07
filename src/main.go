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
func checkErrorReturn(_ interface{}, err error) {
	// Is used if the return is ignored but the error is not
	if err != nil {
		logger.Error(err)
	}
}
