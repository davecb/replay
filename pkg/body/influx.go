package influx

import (
	"context"
	"github.com/davecb/replay/pkg/replay"
	"go.uber.org/zap"
)

// GetInfluxDatum gets a result, normally from influx, but optionally from a replay file.
func GetInfluxDatum(ctx context.Context, logger *zap.SugaredLogger) ([]replay.ConfusionMatrixInfoLog, error) {
	var m []replay.ConfusionMatrixInfoLog
	if replay.Active {
		m, err := replay.Get("thing to grep for")
		if err != nil {
			return m, err
		}
		return m, nil
	}
	// return other stuff, from influx. Uses context.
	var fred replay.ConfusionMatrixInfoLog
	logger.Info("thing to grep for", zap.Any("ConfusionMatrixInfoLog", fred))
	//replay.InfoToJson(fred)
	return m, nil
}
