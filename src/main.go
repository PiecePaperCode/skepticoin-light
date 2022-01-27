package main

import (
	"github.com/wonderivan/logger"
)

func main() {
	PeerEvent.Hello()
	for {
		receive()
	}
}
func checkError(err error) {
	if err != nil {
		logger.Emer(err)
	}
}
func checkErrorReturn(_ interface{}, err error) {
	// Is used if the return is ignored but the error is not
	if err != nil {
		logger.Emer(err)
	}
}
