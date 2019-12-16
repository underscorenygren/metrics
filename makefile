.PHONY: build help test ls docs
RUNNER_DIR := examples
BINARY_DIR := cmd
INSTALL_PATH := $(GOPATH)/bin

## Makefile self-documentation
##############################

# From: https://suva.sh/posts/well-documented-makefiles/#grouped-makefile
BLUESTYLE=\033[36m
BOLDSTYLE=\033[1m
ENDSTYLE=\033[0m
PADDING=25
PADDINGSTR=$(shell printf "%-${PADDING}s" ' ')


HELP_AWK_CMD=BEGIN $\
		{ $\
			FS = ":.*\#\#"; $\
			printf "MAKE COMMANDS\n--------------------\n make $(BLUESTYLE)<command>$(ENDSTYLE)\n" $\
		} $\
		/^[\\%%a-zA-Z_-]+:.*?\#\#/ $\
		{ $\
			printf "$(BLUESTYLE)%-$(PADDING)s$(ENDSTYLE) %s\n", $$1, $$2 $\
		} $\
		/^\#\#@/ $\
		{ $\
			printf "\n$(BOLDSTYLE)%s$(ENDSTYLE)\n", substr($$0, 5) $\
		}

help:
	$(eval WIDTH=90)
	@awk '$(HELP_AWK_CMD)' Makefile | while read line; do \
        if [[ $${#line} -gt $(WIDTH) ]] ; then \
			echo "$$line" | fold -sw$(WIDTH) | head -n1; \
			echo "$$line" | fold -sw$(WIDTH) | tail -n+2 | sed "s/^/  $(PADDINGSTR)/"; \
		else \
			echo "$$line"; \
		fi; done
	@echo "\nTARGETS"
	@echo "--------------------"
	@make ls

#####################################
## End of Makefile self-documentation

# Local

run-%:  ## runs a target locally
	go run $(RUNNER_DIR)/$*/main.go

build-%:  ## builds a target locally
	go build  -ldflags "-X main.version=$(VERSION)" -o build/$* $(RUNNER_DIR)/$*/main.go

# we do this "by hand", because I can't figure out how to properly do it using golang, using `install` or whatnot
install-%:  ## installs a target binary
	go build -ldflags "-X main.version=$(VERSION)" -o $(INSTALL_PATH)/$* $(BINARY_DIR)/$*/main.go

clean-%:  ## cleans up build targets
	rm -f build/$*

# derive targets from folders in runner directory
# suppress errors since one will be empty in docker and vice versa
LS_CMD = ls $(RUNNER_DIR)/ 2> /dev/null
BIN_LS_CMD = ls $(BINARY_DIR)/ 2> /dev/null
TARGET_NAMES := $(shell $(LS_CMD))
BIN_TARGET_NAMES := $(shell $(BIN_LS_CMD))
BUILD_TARGETS := $(addprefix build-,$(TARGET_NAMES))
CLEAN_TARGETS := $(addprefix clean-,$(TARGET_NAMES))
INSTALL_TARGETS := $(addprefix install-,$(BIN_TARGET_NAMES))
VERSION := $(shell cat Version)

build: $(BUILD_TARGETS) ## builds all targets

clean: $(CLEAN_TARGETS) ## cleans all targets

install: $(INSTALL_TARGETS) ## installs all binaries

install-deps: ## install dependencies
	go get

install-test: install-deps ## install prereqs for testing
	pip install localstack

## Docker

docker-install:  ## installs dependencies inside golang docker container
	go mod download

localstack:
	localstack start

# Utilities

ls:  ## lists available build/run targets
	@$(LS_CMD) | cat

test:  ## runs specs
	@ginkgo -r ./...

docs:  ## generates documentation
	docker build -f docs.Dockerfile -t partaj-docs .
	docker run --rm -it -p 6060:6060 partaj-docs

open-docs:
	open http://localhost:6060/pkg/github.com/underscorenygren/partaj/
