
all: minebacktrace

deps:
	git submodule update --init  --recursive
	make -C extern/filecoin-ffi all

minebacktrace: deps
	go build -o backtrace main.go


