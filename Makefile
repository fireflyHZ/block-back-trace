#eg : make miner=f02420
all: profit forward

profit:
	go build -o $(miner) main.go
.PHONY: forward
forward:
	go build -o forward ./forwarding/forward.go
