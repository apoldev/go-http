package logger

type Logger interface {
	Print(args ...interface{})
	Printf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
}
