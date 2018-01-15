// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"os"

	"github.com/evalphobia/logrus_sentry"
	"github.com/sirupsen/logrus"

	. "bitbucket.org/delving/rapid/config"
)


func NewLogger() *logrus.Logger {
	l := logrus.New()
	l.Out = os.Stdout
	l.Level = logrus.DebugLevel
	addSentry(l)
	return l
}

// addSentry add the Sentry logging hook when a DSN is defined in the Config
func addSentry(logger *logrus.Logger) {
	dsn := Config.Logging.SentryDSN
	if dsn != "" {
		logger.WithField("dsn", dsn).Infoln("Adding Sentry logging hook.")
		hook, err := logrus_sentry.NewSentryHook(dsn, []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		})

		if err != nil {
			logger.WithField("dsn", dsn).Fatalln("Unable to start sentry with specified DSN.")
		}
		logger.Hooks.Add(hook)

	}
}
