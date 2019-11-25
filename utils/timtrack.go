package utils

import (
	"fmt"
	"log"
	"time"
)

func TimeTrack(start time.Time, msg_opt... string ) {

	elapsed := time.Since(start)

	// Skip this function, and fetch the PC and file for its parent.
	//pc, _, _, _ := runtime.Caller(1)

	// Retrieve a function object this functions parent.
	//funcObj := runtime.FuncForPC(pc)

	// Regex to extract just the function name (and not the module path).
	//runtimeFunc := regexp.MustCompile('^.*\.(.*)$')
	//name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	msg := ""
	if len(msg_opt) > 0 {
		msg = msg_opt[0]
	}

	log.Println(fmt.Sprintf("%s | Прошло: %s", msg, elapsed))
}