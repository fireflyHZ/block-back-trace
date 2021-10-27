#eg : make miner=f02420
all: profit

deps:
	#git submodule update --init  --recursive
	make -C extern/filecoin-ffi all

profit: deps
	go build -o firefly main.go


