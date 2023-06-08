package log

import (
	"io"
	"log"
	"os"
	"runtime"
	"sync"
)

var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)

	loggers = []*log.Logger{errorLog, infoLog}
	mu      sync.Mutex
)

var (
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)

const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

func init() {
	// 如果是Windows就先关闭颜色打印，后续有解决方案了再添加
	if runtime.GOOS == "windows" {
		errorLog = log.New(os.Stdout, "===[error]=== ", log.LstdFlags|log.Lshortfile)
		infoLog = log.New(os.Stdout, "[info ] ", log.LstdFlags|log.Lshortfile)
		loggers = []*log.Logger{errorLog, infoLog}
		Error = errorLog.Println
		Errorf = errorLog.Printf
		Info = infoLog.Println
		Infof = infoLog.Printf
	}
}

// SetLevel用于给当前日志分等级
// 可用的分级InfoLevel,ErrorLevel,Disabled
func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if ErrorLevel < level {
		errorLog.SetOutput(io.Discard)
	}

	if InfoLevel < level {
		infoLog.SetOutput(io.Discard)
	}
}
