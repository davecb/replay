build:
	cd cmd/example; go build -o /home/davecb/go/bin/replay

run: /home/davecb/go/bin/replay
	/home/davecb/go/bin/replay

replay: /home/davecb/go/bin/replay
	/home/davecb/go/bin/replay --replay ./cmd/example/testdata/onechunk.log
