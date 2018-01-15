package models

import (
	"bitbucket.org/delving/rapid/hub3/logging"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// init added all the logrus hooks
func init() {
	logger = logging.NewLogger()
}
