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
		fmt.Fprintf(os.Stderr, "Log level env variable is not specified. Use INFO")
		level = slog.LevelInfo
	} else {
		err := level.UnmarshalText([]byte(s))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bad log level env variable: '%v'. Use INFO", s)
			level = slog.LevelInfo
		} else {
			fmt.Fprintf(os.Stdout, "App log level: '%v'", s)
		}
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
