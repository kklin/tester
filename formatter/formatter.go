package formatter

import (
	"github.com/Sirupsen/logrus"
)

var Formatter = &logrus.TextFormatter{
	FullTimestamp: true,
}
