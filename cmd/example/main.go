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
	var replayFile string

	flag.StringVar(&replayFile, "replay", "", "file to replay")
	flag.Parse()

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

// FIXME DCB start logger, get good output, then simplify

// setUpLogger creates a sugared logger with caller, timestamp, etc.
func setUpLogger() (*zap.SugaredLogger, error) {
	logger, err := NewLogger(&LoggerConfig{
		Level:       cfg.Level,
		Development: cfg.Development,
		Encoding:    cfg.Encoding,
		OutputPaths: cfg.OutputPaths,
	})
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(logger)
	zap.S().Infof("Successfully initialized logger")

	return logger.Sugar(), nil
}

// LoggerSamplingConfig contains relevant configuration for creating a sampling zap.Logger.
type LoggerSamplingConfig struct {
	Initial    int `toml:"initial" mapstructure:"initial"`
	Thereafter int `toml:"thereafter" mapstructure:"thereafter"`
}

// LoggerConfig contains relevant configuration for creating a zap.Logger.
type LoggerConfig struct {
	Level       string                `toml:"level" mapstructure:"level"`
	OutputPaths []string              `toml:"output_paths" mapstructure:"output_paths"`
	Encoding    string                `toml:"encoding" mapstructure:"encoding"`
	Development bool                  `toml:"development" mapstructure:"development"`
	Sampling    *LoggerSamplingConfig `toml:"sampling" mapstructure:"sampling"`
}

var cfg = &LoggerConfig{
	Level:       "debug",
	Development: true,
	Encoding:    "json",
	OutputPaths: []string{"stderr"},
}

// newZapConfig maps a LoggerConfig into a zap.Config for creating a zap.Logger.
func newZapConfig(cfg *LoggerConfig) *zap.Config {
	var scfg *zap.SamplingConfig
	if cfg.Sampling != nil {
		scfg = &zap.SamplingConfig{
			Initial:    cfg.Sampling.Initial,
			Thereafter: cfg.Sampling.Thereafter,
		}
	}
	zcfg := &zap.Config{
		Encoding:    cfg.Encoding,
		OutputPaths: cfg.OutputPaths,
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
		Development: cfg.Development,
		Sampling:    scfg,
	}

	_ = zcfg.Level.UnmarshalText([]byte(cfg.Level))
	return zcfg
}

func NewLogger(cfg *LoggerConfig) (*zap.Logger, error) {
	return newZapConfig(cfg).Build()
}

// getSettingsFromCommandLineFlags is a minimal example.
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
