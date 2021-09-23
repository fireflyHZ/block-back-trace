package log

import (
	logging "github.com/ipfs/go-log/v2"
)

func Init() error {

	//controller
	//if err := logging.SetLogLevel("order-ctr-log", "DEBUG"); err != nil {
	//	return err
	//}
	//if err := logging.SetLogLevel("reward-ctr-log", "DEBUG"); err != nil {
	//	return err
	//}
	//if err := logging.SetLogLevel("user-ctr-log", "DEBUG"); err != nil {
	//	return err
	//}
	//
	//if err := logging.SetLogLevel("block-ctr-log", "DEBUG"); err != nil {
	//	return err
	//}
	//models
	if err := logging.SetLogLevel("models", "DEBUG"); err != nil {
		return err
	}
	if err := logging.SetLogLevel("lotus-setup", "DEBUG"); err != nil {
		return err
	}
	if err := logging.SetLogLevel("message-log", "DEBUG"); err != nil {
		return err
	}
	if err := logging.SetLogLevel("reward-former-log", "DEBUG"); err != nil {
		return err
	}
	if err := logging.SetLogLevel("user-log", "DEBUG"); err != nil {
		return err
	}
	if err := logging.SetLogLevel("reward-log", "DEBUG"); err != nil {
		return err
	}
	if err := logging.SetLogLevel("block-log", "DEBUG"); err != nil {
		return err
	}
	//logging.SetupLogging(logging.Config{File: "profit.log"})
	return nil
}
