package tool

import (
	"fmt"
	"time"
)

//Logger A new logging structure
type Logger struct {
	Level int
}

func (logger *Logger) Error(s string) {

	if logger.Level > 1 {
		datetime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("%s ERROR: %s \n", datetime, s)
	}
}

//Warning warning
func (logger *Logger) Warning(s string) {
	if logger.Level > 2 {
		datetime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("%s WARNING: %s \n", datetime, s)
	}
}

//Info give out info
func (logger *Logger) Info(s string) {
	if logger.Level > 3 {
		datetime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("%s INFO: %s \n", datetime, s)
	}
}

//Debug debug info
func (logger *Logger) Debug(s string) {
	if logger.Level > 4 {
		datetime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("%s DEBUG: %s\n", datetime, s)
	}
}
func (logger *Logger) write(s string) {

}

// func main() {
// 	logger := &Log{Level: 3}
// 	logger.info("info")
// 	logger.warning("warning")
// 	logger.debug("debug")
// 	logger.error("error")
// }
