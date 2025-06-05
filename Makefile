.PHONY: build test clean install

build:
	go build -o mcp-kokoro

test:
	cd test && ./scripts/test_kokoro.sh

clean:
	rm -f mcp-kokoro

install:
	go install

run:
	./mcp-kokoro