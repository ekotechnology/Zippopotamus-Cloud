date = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
package = github.com/zippopotamus/zippopotamus
flagPath = $(package)/internal
targets = api parser
SHA=$(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=normal)
ifneq ($(GITUNTRACKEDCHANGES),)
	SHA := $(SHA)-dirty
endif
version?=notset
FLAGS = -X '$(flagPath).Version=$(version)' -X '$(flagPath).GitCommit=$(SHA)' -X '$(flagPath).BuildDate=$(date)'
PROGRAM := $(addprefix cmd/zp-,$(cmd))

.PHONY: default
default: $(targets)

.PHONY: format
format:
	gofmt -w internal cmd

include make/bins.mk
include make/docker.mk
include make/static_data.mk
include make/postal_data.mk

.PHONY: clean
clean: clean-admincodes clean-bins clean-data clean-generated-admincodes
