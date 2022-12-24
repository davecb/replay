package main

import (
	"context"
	"flag"
	"fmt"
	influx "github.com/davecb/replay/pkg/body"
	"github.com/davecb/replay/pkg/replay"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	var fred replay.ConfusionMatrixInfoLog
	replay.InfoToJson(fred)
	os.Exit(-1)

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
		// use a file of replay-data for input instead of ifluxDB
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
		datum, err := influx.GetInfluxDatum(i)
		if err != nil {
			return err
		}
		logger.Debug("got %q\n", datum)
	}
	return nil
}

// setUpLogger creates a sugared logger for the example
func setUpLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}

func getSettingsFromCommandLineFlags() (string, error) {
	// load config
	var replay string
	flag.StringVar(&replay, "replay", "", "file to replay")

	flag.Parse()
	return replay, nil
}

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
