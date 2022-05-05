NAME := githubpal
VERSION := $(shell git describe --tags --always)
SRC_DIR := ./cmd/$(NAME)

GO_BUILD := go build -ldflags "-X main.Version=$(VERSION)"

build: clean
	@echo "build $(VERSION)"
	@$(GO_BUILD) -o ./build/$(NAME) $(SRC_DIR)

install:
	@echo "installing to $(GOPATH)/bin"
	@cd $(SRC_DIR) && go install

clean:
	@echo "cleaning artifacts"
	@rm -rf build/ && mkdir build/
