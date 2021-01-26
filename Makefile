all: profit forward

profit:
	go build -o profit-allocation main.go
.PHONY: forward
forward:
	go build -o forward ./forwarding/forward.go

