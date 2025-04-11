package logger

import (
	"context"
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"runtime"
	"time"
)

// Keys used in log fields.
const (
	ErrorKey          = "error"
	SpanIdLogKeyName  = "spanID"
	TraceIdLogKeyName = "traceID"
)

// baseLogger is our global base logger configuration.
var baseLogger *logrus.Logger

// init initializes a Logrus-based logger with JSON output, Debug level,
// no built-in caller info (we add our own).
func init() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	log.SetReportCaller(false)
	log.SetLevel(logrus.DebugLevel)
	baseLogger = log
}

// Log returns a logger instance potentially enriched with Sentry Trace and Span IDs
// extracted from the context if a Sentry span is active.
func Log(ctx context.Context) ILogger {
	entry := logrus.NewEntry(baseLogger) // Start with a base entry for each call
	fields := logrus.Fields{}

	// Attempt to extract Sentry span context
	if ctx != nil {
		if span := sentry.SpanFromContext(getSentryContext(ctx)); span != nil {
			traceID := span.TraceID.String()
			spanID := span.SpanID.String()
			if !isIdNull(traceID) && !isIdNull(spanID) {
				fields[TraceIdLogKeyName] = traceID
				fields[SpanIdLogKeyName] = spanID
			}
		}
	}

	// Return logger with potentially added Sentry fields
	return &Logger{Entry: entry.WithFields(fields), Context: ctx}
}

// Logger is a Logrus-based ILogger.
type Logger struct {
	Entry   *logrus.Entry
	Context context.Context
}

// SetLevel sets the log level for the underlying logger instance.
func (l *Logger) SetLevel(level logrus.Level) ILogger {
	l.Entry.Logger.SetLevel(level)
	return l
}

// Debug logs a message at debug level with optional fields.
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.Entry.WithFields(l.convertToFields(keysAndValues...)).
		WithField("caller", getCallerInfo()).
		Debug(msg)
}

// Info logs a message at info level with optional fields.
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.Entry.WithFields(l.convertToFields(keysAndValues...)).
		WithField("caller", getCallerInfo()).
		Info(msg)
}

// Warn logs a message at warning level with optional fields.
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.Entry.WithFields(l.convertToFields(keysAndValues...)).
		WithField("caller", getCallerInfo()).
		Warn(msg)
}

// Error logs a message at error level with optional fields.
// Sentry capture should happen explicitly where the error is handled using the context-aware hub.
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.captureErrors(msg)
	fields := l.convertToFields(keysAndValues...)
	l.Entry.WithFields(fields).
		WithField("caller", getCallerInfo()).
		Error(msg)
}

// Fatal logs a message at fatal level, then exits.
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.captureErrors(msg)

	fields := l.convertToFields(keysAndValues...)
	l.Entry.WithFields(fields).
		WithField("caller", getCallerInfo()).
		Fatal(msg)
}

// Panic logs a message at panic level, then panics.
func (l *Logger) Panic(msg string, keysAndValues ...interface{}) {
	l.captureErrors(msg)

	fields := l.convertToFields(keysAndValues...)
	l.Entry.WithFields(fields).
		WithField("caller", getCallerInfo()).
		Panic(msg) // Note: Panic triggers a panic
}

// WithValues returns a new ILogger with additional fields added to the entry.
func (l *Logger) WithValues(keysAndValues ...interface{}) ILogger {
	fields := l.convertToFields(keysAndValues...)
	return &Logger{Entry: l.Entry.WithFields(fields), Context: l.Context}
}

// WithError returns a new ILogger that includes the given error in the log context.
func (l *Logger) WithError(err error) ILogger {
	return &Logger{Entry: l.Entry.WithField(ErrorKey, err), Context: l.Context}
}

// convertToFields turns key-value pairs into Logrus fields.
func (l *Logger) convertToFields(keysAndValues ...interface{}) logrus.Fields {
	fields := logrus.Fields{}
	length := len(keysAndValues)
	if length == 0 {
		return fields
	}
	for i := 0; i < length-1; i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			key = fmt.Sprintf("invalid_key_%d", i)
		}
		fields[key] = keysAndValues[i+1]
	}
	if length%2 != 0 {
		key, ok := keysAndValues[length-1].(string)
		if !ok {
			key = fmt.Sprintf("invalid_key_%d", length-1)
		}
		fields[key] = "MISSING_VALUE"
	}
	return fields
}

func (l *Logger) captureErrors(msg string) {
	if l.Context == nil {
		return
	}

	var hub *sentry.Hub
	if h, ok := l.Context.Value(sentry.HubContextKey).(*sentry.Hub); ok {
		hub = h
	} else {
		return
	}

	if msg != "" {
		hub.CaptureException(errors.New(msg))
	}

	// Capture errors in fields
	fields := l.Entry.Data
	for _, value := range fields {
		if err, isError := value.(error); isError {
			hub.CaptureException(err)
		}
	}
}

// getCallerInfo returns file:line and function name for the calling code.
func getCallerInfo() string {
	const skipFrames = 3
	pc, file, line, ok := runtime.Caller(skipFrames)
	if !ok {
		return "unknown:?"
	}
	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = filepath.Base(fn.Name())
	}
	return fmt.Sprintf("%s:%d %s", filepath.Base(file), line, funcName)
}

// isIdNull checks if a trace or span ID string is effectively null (empty or all zeros).
func isIdNull(id string) bool {
	if len(id) == 0 {
		return true
	}
	for _, c := range id {
		if c != '0' {
			return false
		}
	}
	return true
}

func getSentryContext(ctx context.Context) context.Context {
	if sentryCtx, ok := ctx.Value(sentry.HubContextKey).(context.Context); ok {
		return sentryCtx
	}
	return ctx
}

// Interface guard ensures Logger implements ILogger at compile time.
var _ ILogger = (*Logger)(nil)
