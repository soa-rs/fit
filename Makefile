SOURCES := $(wildcard cmd/server/*.go)
BINARY := soarsfit

.PHONY: run build clean

run:
	go run $(SOURCES)

build:
	go build -o $(BINARY) $(SOURCES)

clean:
	rm -f $(BINARY)
