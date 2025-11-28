package mlog

var (
	logObj = NewLogger(LOG_DEBUG, &StdWriter{})
)

func Init(level string, lpath string, lname string) {
	logObj.Close()
	logObj = NewLogger(level, NewLogWriter(lpath, lname))
}

func Close() {
	logObj.Close()
}

func Tracef(format string, args ...any) {
	logObj.Trace(1, format, args...)
}

func Debugf(format string, args ...any) {
	logObj.Debug(1, format, args...)
}

func Warnf(format string, args ...any) {
	logObj.Warn(1, format, args...)
}

func Infof(format string, args ...any) {
	logObj.Info(1, format, args...)
}

func Errorf(format string, args ...any) {
	logObj.Error(1, format, args...)
}

func Fatalf(format string, args ...any) {
	logObj.Fatal(1, format, args...)
}

func Trace(skip int, format string, args ...any) {
	logObj.Trace(skip+1, format, args...)
}

func Debug(skip int, format string, args ...any) {
	logObj.Debug(skip+1, format, args...)
}

func Warn(skip int, format string, args ...any) {
	logObj.Warn(skip+1, format, args...)
}

func Info(skip int, format string, args ...any) {
	logObj.Info(skip+1, format, args...)
}

func Error(skip int, format string, args ...any) {
	logObj.Error(skip+1, format, args...)
}

func Fatal(skip int, format string, args ...any) {
	logObj.Fatal(skip+1, format, args...)
}
