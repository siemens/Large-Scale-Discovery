/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package log

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
)

// ZapLogger is a small wrapper for zap's SugaredLogger, that implements our Logger interface
type ZapLogger struct {
	*zap.SugaredLogger
	closeOnce *sync.Once
	closeFns  []func() error
}

func NewZapLogger(logger *zap.SugaredLogger, closeFns ...func() error) Logger {
	return &ZapLogger{
		logger,
		&sync.Once{},
		closeFns,
	}
}

// Warningf is a sample renaming for the zap loggers Warnf method
func (l *ZapLogger) Warningf(template string, args ...interface{}) {
	l.SugaredLogger.Warnf(template, args...)
}

func (l *ZapLogger) Tagged(tag string) Logger {
	return &ZapLogger{l.SugaredLogger.Named(tag), l.closeOnce, l.closeFns}
}

func (l *ZapLogger) Close() (errs error) {

	// Sync the logger before shutting it down
	err := l.SugaredLogger.Sync()
	if err != nil {
		errs = err
	}

	// Call the close functions of the different cores
	l.closeOnce.Do(func() {
		for _, closeFn := range l.closeFns {
			errClose := closeFn()
			if errClose != nil {
				if errs == nil {
					errs = errClose
					continue
				}
				errs = fmt.Errorf("%w; %s", errs, errClose)
			}
		}
	})
	return
}
