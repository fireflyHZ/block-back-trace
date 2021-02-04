package main

import (
	logging "github.com/ipfs/go-log/v2"
)

func initlog() error {
	logging.SetupLogging(logging.Config{File: "forward.log"})
	//controller
	if err := logging.SetLogLevel("forwardLog", "Info"); err != nil {
		return err
	}
	return nil
}
