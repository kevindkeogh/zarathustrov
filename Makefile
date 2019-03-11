
PKGS := $(shell ls ./cmd)

.PHONY: build fmt clean install
build: fmt
	mkdir -p bin
	go build -x -v -o ./bin/zarathustrov

install:
	go install

fmt:
	@for package in $(PKGS) ; \
	do \
		go fmt ./cmd/$$package ; \
	done

clean:
	@rm -rf ./bin
