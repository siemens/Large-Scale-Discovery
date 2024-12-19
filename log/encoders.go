/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package log

import (
	"fmt"
	"go.uber.org/zap/zapcore"
	"time"
)

const TimestampFormat = "2006-01-02 15:04:05.000 -0700"

func NameTagEncoder(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%-16s", fmt.Sprintf("[%s]", loggerName)))
}

func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}
	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, TimestampFormat)
		return
	}
	enc.AppendString(t.Format(TimestampFormat))
}
