.PHONY: build install clean

build:
	go build -o nl2cmd main.go

install: build
	sudo mv nl2cmd /usr/local/bin/

clean:
	rm -f nl2cmd
