package log

import "log"

func Log(op string, err error) {
	log.Printf("op: %s, err:%s", op, err.Error())
}
