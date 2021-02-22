#eg : make miner=f02420
all: profit forward

deps:
	git submodule update --init --recursive
	make -C extern/filecoin-ffi all

profit: deps
	go build -o $(miner) main.go

forward: deps
	go build -o forward ./forwarding/forward.go ./forwarding/log.go
