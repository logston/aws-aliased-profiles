build:
	go build

install: build
	rm -rf /usr/local/bin/aws-aliased-profiles
	cp aws-aliased-profiles /usr/local/bin/

all: build
