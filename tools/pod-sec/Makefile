.PHONY: build install clean test

BUILD_DIR := $(shell pwd)
OUTPUT := pod-sec

build:
	go build -o $(OUTPUT) main.go

install: build
	cp $(OUTPUT) /usr/local/bin/
	cp tool-mapping.json /usr/local/share/pod-sec/
	@echo "Installation complete. Add to your shell rc:"
	@echo "  source $(BUILD_DIR)/pod-sec.sh"

uninstall:
	rm -f /usr/local/bin/$(OUTPUT)
	rm -rf /usr/local/share/pod-sec/
	@echo "Uninstallation complete. Remove the source line from your shell rc."

clean:
	rm -f $(OUTPUT)

test:
	go build -o /dev/null main.go

run-example:
	./$(OUTPUT) nmap -sV 127.0.0.1

.DEFAULT_GOAL := build
