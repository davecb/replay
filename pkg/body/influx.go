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
	// return other stuff, from influx. Uses context. FIXME DCB
	//{"level":"DEBUG",
	//"timestamp":"2023-01-01T16:38:57.715-0500",
	//"caller":"runner/runner.go:184",
	//"message":"inspecting distribution signal",
	//"run_id":"1672609137356",
	//"job_name":"Regional EMEA Promotion",
	//"split_type":"emea",
	//"DistributionSignal":{
	//	"DistributionID":"20230101T13",
	//	"CappedDSPPriceRetention":{
	//		"RevenueConfusionMatrix":[{
	//			"SumTruePositiveRevenue":28,
	//			"SumFalseNegativeRevenue":8.842105263157897,
	//			"SumTrueNegativeRevenue":0,
	//			"SumFalsePositiveRevenue":0,
	//			"ControlRate":1,
	//			"VdcID":""
	//		}],
	//		"NumMinutes":14},
	//	"UncappedDSPPriceRetention":{
	//		"RevenueConfusionMatrix":[{
	//			"SumTruePositiveRevenue":28,
	//			"SumFalseNegativeRevenue":0,
	//			"SumTrueNegativeRevenue":0,
	//			"SumFalsePositiveRevenue":0,
	//			"ControlRate":1,
	//			"VdcID":""
	//			}],
	//		"NumMinutes":14}},
	//		"UncappedDSPCalculatedPriceRetention":1,
	//		"CappedDSPCalculatedPriceRetention":0.7599999999999999}
	logger.Info("thing to grep for", zap.Int("InterestingVariable", 42))
	//replay.InfoToJson(fred)
	return m, nil
}
