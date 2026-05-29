.PHONY: build run test clean

build:
	go build -o vibeview.exe .

run: build
	./vibeview.exe

test:
	go test ./internal/... -v

clean:
	rm -f vibeview.exe
