build:
	go build

install: build
	cp aws-aliased-profiles /usr/local/bin/

all: build
