package app

import (
	"fmt"
	"group_by_file_types/internal/process"
	"log/slog"
	"os"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.SetDefault(logger)
}

func Run() {
	if len(os.Args) < 2 {
		fmt.Println("Enter the path to the folder where files should be grouped: <path>")
		return
	}
	if len(os.Args) > 2 {
		fmt.Println("Too many arguments. Please enter only the path to the folder.")
		return
	}

	processer := process.NewProcesser(os.Args[1])
	process.Processing(processer)
}
