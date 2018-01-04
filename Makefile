
PKGS := $(shell ls ./cmd)

.PHONY: build fmt clean
build: fmt dep
	@mkdir -p bin
	@for package in $(PKGS) ; \
	do \
		go build -x -v -o ./bin/$$package ./cmd/$$package/*.go ; \
	done

fmt:
	@for package in $(PKGS) ; \
	do \
		go fmt ./cmd/$$package/*.go ; \
	done

dep:
	@dep ensure

clean:
	@rm -rf ./bin
