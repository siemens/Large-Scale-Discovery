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
	"github.com/siemens/ZapSmtp/cores"
	"github.com/siemens/ZapSmtp/smtp"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"sync"
)

var globalLogger Logger
var initOnce sync.Once

type Logger interface {
	io.Closer
	Sync() error

	Tagged(string) Logger
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warningf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

// GetLogger returns the same single global logger instance to all callers
func GetLogger() Logger {
	return globalLogger
}

// InitGlobalLogger initializes a GLOBAL logger based on a given configuration struct
func InitGlobalLogger(conf Settings) (Logger, error) {

	var err error
	initOnce.Do(func() {
		c := make([]zapcore.Core, 0, 3)
		var closeFns []func() error

		// Create the different cores depending on the config. Anonymous function so we can handle errors better
		err = func() error {
			if conf.Console != nil && conf.Console.Enabled {
				core, errCore := initConsoleCore(conf.Console)
				if errCore != nil {
					return errCore
				}
				c = append(c, core)
			}

			if conf.File != nil && conf.File.Enabled {
				core, closeCoreFn, errCore := initFileCore(conf.File)
				if errCore != nil {
					return errCore
				}
				c = append(c, core)
				closeFns = append(closeFns, closeCoreFn)
			}

			if conf.Smtp != nil && conf.Smtp.Enabled {
				core, closeCoreFn, errCore := initSmtpCore(conf.Smtp)
				if errCore != nil {
					return errCore
				}
				c = append(c, core)
				closeFns = append(closeFns, closeCoreFn)
			}

			return nil
		}()

		if err != nil {
			for _, f := range closeFns {
				errF := f()
				if errF != nil {
					err = multierr.Append(err, errF)
				}
			}
			return
		}

		// Tee all the cores together
		tee := zapcore.NewTee(c...)

		// Set the global logger
		globalLogger = NewZapLogger(zap.New(tee).Sugar(), closeFns...)
	})

	return globalLogger, err
}

// CloseGlobalLogger will call the Close method of the global logger
func CloseGlobalLogger() error {
	if globalLogger != nil {
		return globalLogger.Close()
	}

	return nil
}

// WrappedWriteSyncer is a helper struct implementing zapcore.WriteSyncer to
// wrap a standard os.Stdout handle, giving control over the WriteSyncer's
// Sync() function. Sync() results in an error on Windows in combination with
// os.Stdout ("sync /dev/stdout: The handle is invalid."). WrappedWriteSyncer
// simply does nothing when Sync() is called by Zap.
type WrappedWriteSyncer struct {
	file *os.File
}

func (mws WrappedWriteSyncer) Write(p []byte) (n int, err error) {
	return mws.file.Write(p)
}
func (mws WrappedWriteSyncer) Sync() error {
	return nil
}

// InitConsoleCore creates a new core for logging to the console according to the provided configuration
func initConsoleCore(conf *ConsoleHandler) (zapcore.Core, error) {

	ws := zapcore.Lock(WrappedWriteSyncer{os.Stdout})

	// Create the encoder. We prefer to have a custom Name (/Tag) Encoder
	encConf := zap.NewDevelopmentEncoderConfig()
	encConf.EncodeName = NameTagEncoder
	encConf.EncodeTime = CustomTimeEncoder
	enc := zapcore.NewConsoleEncoder(encConf)

	// Create the core
	return zapcore.NewCore(enc, ws, conf.Level), nil
}

// InitFileCore creates a new core for logging to a file according to the provided configuration
func initFileCore(conf *FileHandler) (zapcore.Core, func() error, error) {

	w := &lumberjack.Logger{
		Filename:   conf.Path,
		MaxSize:    conf.SizeMb, // megabytes
		MaxBackups: conf.History,
		MaxAge:     28, // days
	}

	ws := zapcore.AddSync(w)

	// Create the encoder. We prefer to have a custom Name (/Tag) Encoder
	encConf := zap.NewDevelopmentEncoderConfig()
	encConf.EncodeName = NameTagEncoder
	encConf.EncodeTime = CustomTimeEncoder
	enc := zapcore.NewConsoleEncoder(encConf)

	// Create the core
	return zapcore.NewCore(enc, ws, conf.Level), w.Close, nil
}

func initSmtpCore(conf *SmtpHandler) (zapcore.Core, func() error, error) {

	// Use a sink as it performs a bit better
	wsc, err := smtp.NewWriteSyncCloser(
		conf.Connector.Server,
		conf.Connector.Port,
		conf.Connector.Username,
		conf.Connector.Password,
		conf.Connector.Subject,
		conf.Connector.Sender,
		conf.Connector.Recipients,
		conf.Connector.OpensslPath,
		conf.Connector.SignatureCertPath,
		conf.Connector.SignatureKeyPath,
		conf.Connector.EncryptionCertPaths,
		conf.Connector.TempDir,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to initilialize stmp write syncer: %s", err)
	}

	// Create the encoder. We prefer to have a custom Name (/Tag) Encoder
	encConf := zap.NewDevelopmentEncoderConfig()
	encConf.EncodeName = NameTagEncoder
	encConf.EncodeTime = CustomTimeEncoder

	enc := zapcore.NewConsoleEncoder(encConf)

	core, err := cores.NewDelayedCore(conf.Level, enc, wsc, conf.LevelPriority, conf.Delay, conf.DelayPriority)
	if err != nil {
		err = fmt.Errorf("unable to initilialize delayed core: %s", err)

		// Close the newly created files
		errC := wsc.Close()
		if errC != nil {
			err = fmt.Errorf("%w; %s", err, errC)
		}

		return nil, nil, err
	}

	return core, wsc.Close, nil
}
