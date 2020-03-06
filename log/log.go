package log

import (
	kitlog "github.com/go-kit/kit/log"
	kitlogrus "github.com/go-kit/kit/log/logrus"
	"github.com/sirupsen/logrus"
	"os"
)

var (
	Logger kitlog.Logger
	Logrus logrus.Logger
)

func init() {
	Logrus := logrus.New()
	//设置控制台输出
	Logrus.Out = os.Stdout
	Logrus.Level = logrus.TraceLevel
	Logrus.Formatter = &logrus.TextFormatter{TimestampFormat: "02-01-2006 15:04:05", FullTimestamp: true, ForceColors: true}
	Logger = kitlogrus.NewLogrusLogger(Logrus)
}
