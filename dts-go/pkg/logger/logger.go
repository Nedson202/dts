package logger

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	once sync.Once
	Log  zerolog.Logger
)

func Init() {
	once.Do(func() {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		Log = zerolog.New(output).With().Timestamp().Caller().Logger()

		// Set global logger
		log.Logger = Log
	})
}

func Debug() *zerolog.Event {
	return Log.Debug()
}

func Info() *zerolog.Event {
	return Log.Info()
}

func Warn() *zerolog.Event {
	return Log.Warn()
}

func Error() *zerolog.Event {
	return Log.Error()
}

func Fatal() *zerolog.Event {
	return Log.Fatal()
}

func Panic() *zerolog.Event {
	return Log.Panic()
}

func Printf(format string, v ...interface{}) {
	Log.Printf(format, v...)
}

func Println(v ...interface{}) {
	Log.Print(v...)
}

func Fatalf(format string, v ...interface{}) {
	Log.Fatal().Msgf(format, v...)
}
