package replay

import (
	"bufio"
	"encoding/json"
	"go.uber.org/zap"
	"log"
	"os"
	"strings"
)

/*
 * replayInflux is a service that accepts a file of json logs and replays them on a
 * development or test version of the QA program. It is used for replaying actual
 * events, either as a regression/integration test, and as a debugging tool.
 */

// ConfusionMatrixInfoLog is the data we need for automated QA, in the form of a json log message
type ConfusionMatrixInfoLog struct {
	Level            string `json:"level"`     // INFO
	Timestamp        string `json:"timestamp"` // 2022-10-31T16:10:41.384Z
	Message          string `json:"message"`   // "thing to grep for goes here"
	SumTruePositive  int    `json:"SumTruePositive"`
	SumFalsePositive int    `json:"SumFalsePositive"`
	SumTrueNegative  int    `json:"SumTrueNegative"`
	SumFalseNegative int    `json:"SumFalseNegative"`
	NumMinutes       int    `json:"NumMinutes"`
}

var eg = ConfusionMatrixInfoLog{
	Level:            "INFO",
	Timestamp:        "2022-10-31T16:10:41.384Z",
	Message:          "thing to grep for goes here",
	SumTruePositive:  1,
	SumFalsePositive: 2,
	SumTrueNegative:  3,
	SumFalseNegative: 4,
	NumMinutes:       5,
}

// replayFile is a file from which to read json log lines
var replay struct {
	file     *os.File       // the file we're reading
	scanner  *bufio.Scanner // effectively, a read pointer into it
	pushback string         // the last line read, if unprocessed
	logger   *zap.SugaredLogger
}

// Active tells the caller if they should call Get()
var Active bool

// Open opens the file passed to it, for later use.
func Open(file string, logger *zap.SugaredLogger) error {
	var err error

	replay.file, err = os.Open(file)
	if err != nil {
		return err
	}
	replay.scanner = bufio.NewScanner(replay.file)
	replay.logger = logger
	replay.logger.Infof("replay started, from file %q\n", file)
	replay.logger.Debugf("replay struct = %#v\n", replay)
	Active = true
	return nil
}

func Get(kindWanted string) ([]ConfusionMatrixInfoLog, error) {
	var result ConfusionMatrixInfoLog
	var m []ConfusionMatrixInfoLog
	var distributionId, kindGot string
	var line string
	var err error

	replay.logger.Debugf("replay get called with %q\n", kindWanted)
	for {
		if replay.pushback != "" {
			// get a line from the pushback buffer
			line = replay.pushback
			replay.pushback = ""
		} else {
			// scan another line
			ok := replay.scanner.Scan()
			if !ok {
				// it must be eof
				break
			}
			err = replay.scanner.Err()
			if err != nil {
				// Try skipping this line, hoping that works. Cross fingers!
				continue
			}
			line = replay.scanner.Text()
			// Yay, we have a new line
		}

		log.Printf("line = %q, err = %q\n", line, err)
		if !strings.Contains(line, "thing to grep for") {
			// only look at parsed retention signals
			continue
		}

		result, err = jsonToConfusionsMatrixInfo(line)
		if err != nil {
			// Try skipping this line, hoping that works. I expect it won't, and
			// we'll have to update the code.
			continue
		}
		log.Printf("result[%v] = %v\n", distributionId, result)

		// My working example had two kinds of data, so it needed to handle both.
		// Frtunately they came in batches of one and then the other, so I could treat\
		// them like a c-language parsing problem and use ungetc()
		if kindWanted == kindGot {
			// It's the right kind of data, save it
			m = append(m, result)
		} else {
			// It's the wrong kind, simulate ungetc() and return the array
			replay.pushback = line
			return m, nil
		}
	}
	// postcondition: we hit eof in Scan() and exited the loop, return what we have & EOF
	return m, nil
}

// Close closes the test-data file
func Close() {
	// desirable but not actually mandatory
	if replay.file != nil {
		_ = replay.file.Close()
	}
}

func jsonToConfusionsMatrixInfo(line string) (ConfusionMatrixInfoLog, error) {
	var input ConfusionMatrixInfoLog

	err := json.Unmarshal([]byte(line), &input)
	if err != nil {
		log.Printf("Warning: input format has changed, code changes required. line = %q\n",
			line)
		// Generate a new struct from the line and compare them
		return input, err
	}
	return input, nil
}

func InfoToJson(s ConfusionMatrixInfoLog) {
	j, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	log.Printf("json=%s\n", j)
}
