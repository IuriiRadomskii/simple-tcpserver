package logger

import (
	"fmt"
	"log/slog"
	"os"
)

var LOG *slog.Logger = logger()

func logger() *slog.Logger {
	var level slog.Level
	s := os.Getenv("LOG_LEVEL")
	if s == "" {
		fmt.Fprintf(os.Stderr, "Log level env variable is not specified. Use DEBUG\n\n")
		level = slog.LevelDebug
	} else {
		err := level.UnmarshalText([]byte(s))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bad log level env variable: '%v'. Use INFO\n", s)
			level = slog.LevelInfo
		} else {
			fmt.Fprintf(os.Stdout, "App log level: '%v'\n", s)
		}
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
