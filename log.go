package log

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	Log *logrus.Logger
)

func Logger(isOutputConsole bool, isWriteFile bool, path ...string) (*logrus.Logger, error) {
	Log = logrus.New()
	Log.Formatter = &logrus.TextFormatter{}
	if isWriteFile {
		Path := path[0]
		if !isExists(Path) {
			err := os.Mkdir(Path, 0755)
			if err != nil {
				return Log, err
			}
		}
		hook := newLfsHook(filepath.Join(Path, "log"), 0, 5)
		Log.AddHook(hook)
	}
	Log.SetFormatter(formatter(true))
	Log.SetReportCaller(true)
	if !isOutputConsole {
		Log.SetOutput(ioutil.Discard)
	}
	return Log, nil
}

func formatter(isConsole bool) *nested.Formatter {
	fmtter := &nested.Formatter{
		HideKeys:        true,
		TimestampFormat: "2006-01-02 15:04:05",
		CallerFirst:     true,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			funcInfo := runtime.FuncForPC(frame.PC)
			if funcInfo == nil {
				return "error during runtime.FuncForPC"
			}
			fullPath, line := funcInfo.FileLine(frame.PC)
			return fmt.Sprintf(" [%v:%v]", filepath.Base(fullPath), line)
		},
	}
	if isConsole {
		fmtter.NoColors = false
	} else {
		fmtter.NoColors = true
	}
	return fmtter
}

func newLfsHook(logName string, rotationTime time.Duration, leastDay uint) logrus.Hook {
	writer, err := rotatelogs.New(
		logName+".%Y%m%d%H%M%S",
		// 日志周期(默认每86400秒/一天旋转一次)
		//rotatelogs.WithRotationTime(rotationTime),
		// 清除历史 (WithMaxAge和WithRotationCount只能选其一)
		//rotatelogs.WithMaxAge(time.Hour*24*7), //默认每7天清除下日志文件
		rotatelogs.WithRotationCount(leastDay), //只保留最近的N个日志文件
	)
	if err != nil {
		panic(err)
	}

	// filename for each level
	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
		//}, &logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	}, formatter(false))

	return lfsHook
}

func isExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
