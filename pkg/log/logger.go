/*
Copyright (c) 2023 Nordix Foundation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"context"
	"fmt"
	golog "log"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger The global logger.
//
//nolint:gochecknoglobals
var Logger logr.Logger

// FromContextOrGlobal return a logger from the passed context or the
// global logger.
func FromContextOrGlobal(ctx context.Context) logr.Logger {
	if logger, err := logr.FromContext(ctx); err == nil {
		return logger
	}

	return Logger
}

// New returns a new logger. The level may be "DEBUG" (V(1)) or "TRACE" (V(2)),
// any other string (e.g. "") is interpreted as "INFO" (V(0)). On first call
// the global Logger is set.
func New(name string, level int) logr.Logger {
	logger := newLogger(level).WithName(name)

	once.Do(func() {
		Logger = logger
	})

	return logger
}

//nolint:gochecknoglobals
var once sync.Once

// Fatal log the message using the passed logger and terminate.
func Fatal(logger logr.Logger, msg string, keysAndValues ...any) {
	if z := zapLogger(logger); z != nil {
		z.Sugar().Fatalw(msg, keysAndValues...)
	} else {
		// Fallback to go default
		golog.Fatal(msg, keysAndValues)
	}
}

// Called before "main()". Pre-set a global logger.
//
//nolint:gochecknoinits
func init() {
	Logger = newLogger(0).WithName("Meridio")
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.999-07:00"))
}

func levelEncoder(lvl zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch lvl {
	case zapcore.InfoLevel:
		enc.AppendString("info")
	case zapcore.WarnLevel:
		enc.AppendString("warning")
	case zapcore.ErrorLevel:
		enc.AppendString("error")
	case zapcore.DPanicLevel:
		enc.AppendString("critical")
	case zapcore.PanicLevel:
		enc.AppendString("critical")
	case zapcore.FatalLevel:
		enc.AppendString("critical")
	case zapcore.DebugLevel:
		enc.AppendString("debug")
	case zapcore.InvalidLevel:
		enc.AppendString("debug")
	default:
		enc.AppendString("debug")
	}
}

func newLogger(level int) logr.Logger {
	lvl := level * -1

	zapConfig := zap.NewProductionConfig()

	zapConfig.Level = zap.NewAtomicLevelAt(zapcore.Level(lvl))
	zapConfig.DisableStacktrace = true
	zapConfig.DisableCaller = true
	zapConfig.EncoderConfig.NameKey = "service_id"
	zapConfig.EncoderConfig.LevelKey = "severity"
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.MessageKey = "message"
	// zc.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder (almost works)
	zapConfig.EncoderConfig.EncodeTime = timeEncoder
	zapConfig.EncoderConfig.EncodeLevel = levelEncoder
	zapConfig.Encoding = "json"
	zapConfig.Sampling = nil
	zapConfig.OutputPaths = []string{"stdout"}

	z, err := zapConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("Can't create a zap logger (%v)?", err))
	}

	return zapr.NewLogger(z.With(
		zap.String("version", "1.0.0"), zap.Namespace("extra_data")))
}

// zapLogger returns the underlying zap.Logger.
// NOTE; If exported this breaks the use of different log implementations!
func zapLogger(logger logr.Logger) *zap.Logger {
	if underlier, ok := logger.GetSink().(zapr.Underlier); ok {
		return underlier.GetUnderlying()
	}

	return nil
}
