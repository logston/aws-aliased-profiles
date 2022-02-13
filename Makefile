build:
	go build

install: build
	mkdir -p ~/.local/bin
	rm -rf ~/.local/bin/aws-aliased-profiles
	cp ./aws-aliased-profiles ~/.local/bin/

all: install
