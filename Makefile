ALL_DIRS=$(shell find . \( -path ./.git \) -prune -o -type d -print)
GO_PKGS=$(shell go list ./...)
GO_FILES=$(foreach dir, $(ALL_DIRS), $(wildcard $(dir)/*.go))
PROTO_FILES=$(sort $(wildcard sfproto/*.proto))

ifeq ("$(CIRCLECI)", "true")
	CI_SERVICE = circle-ci
endif

all: test

lint:
	@golint ./... | grep -v '^sfproto\/signalfx\.pb\.go:' || true
	@go vet ./...

test: $(GO_FILES) sfproto/signalfx.pb.go
	go test -v -race ./...

coverage: .acc.out

coveralls: .coveralls-stamp

clean:
	@rm -f \
		./.acc.out \
		./.coveralls-stamp

.acc.out: $(GO_FILES) sfproto/signalfx.pb.go
	@echo "mode: set" > .acc.out
	@for pkg in $(GO_PKGS); do \
		cmd="go test -v -coverprofile=profile.out $$pkg"; \
		eval $$cmd; \
		if test $$? -ne 0; then \
			exit 1; \
		fi; \
		if test -f profile.out; then \
			cat profile.out | grep -v "mode: set" | grep -v "\.pb\.go:" >> .acc.out; \
		fi; \
	done
	@rm -f ./profile.out

.coveralls-stamp: .acc.out
	@if [ -n "$(COVERALLS_REPO_TOKEN)" ]; then \
		goveralls -v -coverprofile=.acc.out -service $(CI_SERVICE) -repotoken $(COVERALLS_REPO_TOKEN); \
	fi
	@touch .coveralls-stamp

sfproto/signalfx.pb.go: $(PROTO_FILES)
	protoc --go_out=. $(PROTO_FILES)
	@sed -i '/^func.*DataPoint).*String() string/d' sfproto/signalfx.pb.go

.PHONY: all lint test coverage coveralls clean
