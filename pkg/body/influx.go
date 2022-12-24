package influx

import "github.com/davecb/replay/pkg/replay"

// GetInfluxDatum gets a result, normally from influx, but optionally from a replay file.
// the boolean is EOF
func GetInfluxDatum(i int) ([]replay.ConfusionMatrixInfoLog, error) {
	var m []replay.ConfusionMatrixInfoLog
	if replay.Active {
		m, err := replay.Get("thing to grep for")
		if err != nil {
			return m, err
		}
	}
	// return other stuff, from influx
	return m, nil
}
