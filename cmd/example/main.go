package main

import (
	"context"
	"flag"
	"fmt"
	influx "github.com/davecb/replay/pkg/body"
	"github.com/davecb/replay/pkg/replay"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	replayFile, err := getSettingsFromCommandLineFlags()
	if err != nil {
		panic(err)
	}

	logger, err := setUpLogger()
	if err != nil {
		fmt.Printf("failed to create logger: %s\n", err.Error())
		panic(err)
	}

	if replayFile != "" {
		// use a file of replay-data for input instead of influxDB
		err := replay.Open(replayFile, logger)
		if err != nil {
			panic(err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	go waitForExitSignal(cancel, signals)

	err = runJob(ctx, replayFile, logger)
	if err != nil {
		panic(err)
	}
}

// runJob is where the real code of your application goes
func runJob(ctx context.Context, replayFile string, logger *zap.SugaredLogger) error {
	for i := 0; i < 10; i++ {
		datum, err := influx.GetInfluxDatum(ctx, logger)
		if err != nil {
			return err
		}
		logger.Debug("got %q\n", datum)
	}
	return nil
}

// setUpLogger creates a sugared logger with caller, timestamp, etc.
func setUpLogger() (*zap.SugaredLogger, error) {
	cfg := &zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "timestamp",
			LevelKey:     "level",
			CallerKey:    "caller",
			MessageKey:   "message",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
			LineEnding:   zapcore.DefaultLineEnding,
		},
	}
	_ = cfg.Level.UnmarshalText([]byte("debug"))

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

// getSettingsFromCommandLineFlags is a miminal exa mple.
func getSettingsFromCommandLineFlags() (string, error) {
	var replay string

	flag.StringVar(&replay, "replay", "", "file to replay")
	flag.Parse()
	return replay, nil
}

// waitForExitSignal calls cancel() if someone hits ^C
func waitForExitSignal(cancel context.CancelFunc, signals chan os.Signal) {
	var sig os.Signal
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig = <-signals
	sigs := map[os.Signal]string{
		syscall.SIGINT:  "SIGINT",
		syscall.SIGTERM: "SIGTERM",
	}
	log.Printf("got %s, closing gracefully", sigs[sig])
	cancel()
}
