.DEFAULT_GOAL := help
IP_ADDRESS := $(shell ifconfig en0 | grep inet | awk '$$1=="inet" {print $$2}')

.PHONY: vm

ifeq ($(VM_DEBUG),true)
    GO_TAGS = -tags vm_debug
    VM_TARGET = debug
else
    GO_TAGS =
    VM_TARGET = all
endif

ifeq ($(shell uname -s),Darwin)
	export CGO_LDFLAGS=-framework Foundation -framework SystemConfiguration
endif

rustdeps: vm core-rust compiler

juno: rustdeps ## compile
	@mkdir -p build
	@go build $(GO_TAGS) -a -ldflags="-X main.Version=$(shell git describe --tags)" -o build/juno ./cmd/juno/

juno-cached:
	@mkdir -p build
	@go build $(GO_TAGS) -ldflags="-X main.Version=$(shell git describe --tags)" -o build/juno ./cmd/juno/

vm:
	$(MAKE) -C vm/rust $(VM_TARGET)

core-rust:
	$(MAKE) -C core/rust $(VM_TARGET)

compiler:
	$(MAKE) -C starknet/rust $(VM_TARGET)

generate: ## generate
	mkdir -p mocks
	go generate ./...

clean-testcache:
	go clean -testcache

test: clean-testcache rustdeps ## tests
	go test $(GO_TAGS) ./...

test-cached: rustdeps ## tests with existing cache
	go test $(GO_TAGS) ./...

test-race: clean-testcache rustdeps
	go test $(GO_TAGS) ./... -race

benchmarks: rustdeps ## benchmarking
	go test $(GO_TAGS) ./... -run=^# -bench=. -benchmem

test-cover: rustdeps ## tests with coverage
	mkdir -p coverage
	go test $(GO_TAGS) -coverpkg=./... -coverprofile=coverage/coverage.out -covermode=atomic ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html

install-deps: | install-gofumpt install-mockgen install-golangci-lint## install some project dependencies

install-gofumpt:
	go install mvdan.cc/gofumpt@latest

install-mockgen:
	go install go.uber.org/mock/mockgen@latest

install-golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.2

lint:
	@which golangci-lint || make install-golangci-lint
	golangci-lint run

tidy: ## add missing and remove unused modules
	 go mod tidy

format: ## run go formatter
	gofumpt -l -w .

clean: ## clean project builds
	$(MAKE) -C vm/rust clean
	$(MAKE) -C core/rust clean
	$(MAKE) -C starknet/rust clean
	@rm -rf ./build

help: ## show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

feedernode: juno-cached
	./build/juno \
	--network=sepolia \
	--log-level=info \
	--db-path=./p2p-dbs/feedernode \
	--p2p \
	--p2p-feeder-node \
	--p2p-addr=/ip4/0.0.0.0/tcp/7777 \
	--p2p-private-key="5f6cdc3aebcc74af494df054876100368ef6126e3a33fa65b90c765b381ffc37a0a63bbeeefab0740f24a6a38dabb513b9233254ad0020c721c23e69bc820089" \
	--metrics-port=9090

node1: juno-cached
# todo remove rm before merge
	rm -rf ./p2p-dbs/node1/ && \
	./build/juno \
	--network=sepolia \
	--log-level=info \
	--metrics \
	--db-path=./p2p-dbs/node1 \
	--p2p \
	--p2p-peers=/ip4/$(IP_ADDRESS)/tcp/7777/p2p/12D3KooWLdURCjbp1D7hkXWk6ZVfcMDPtsNnPHuxoTcWXFtvrxGG \
	--p2p-addr=/ip4/0.0.0.0/tcp/7778 \
	--p2p-private-key="beffe878f402f6d39dd56acffe726dec3dcbffc8fb359c79351a75b382ba070b99d378d539994aa276020ff1fa624ab08419b9ac311925ea41eb7354c2214113" \
	--metrics-port=9091

#	--p2p-peers=/ip4/127.0.0.1/tcp/7778/p2p/12D3KooWDQVMmK6cQrfFcWUoFF8Ch5vYegfwiP5Do2SFC2NAXeBk \

node2:
	./build/juno \
	--network=sepolia \
	--log-level=info \
	--db-path=./p2p-dbs/node2 \
	--p2p \
	--p2p-peers=/ip4/$(IP_ADDRESS)/tcp/7778/p2p/12D3KooWLAqaDSStd48xxHSYpqNoDwNQakBWnyfwoYvtUBUBtBAi \
	--p2p-private-key="f5a41dfde564a047d0ee07943bbefe86c477e2f8c592f22081c28b1756d1f19c7f7e690afa3cc123cde2059ea98fe2a218e475759affe833da73854a45b0853a" \
	--metrics-port=9092

node3:
	./build/juno \
	--network=sepolia \
	--log-level=info \
	--db-path=./p2p-dbs/node3 \
	--p2p \
	--p2p-peers=/ip4/$(IP_ADDRESS)/tcp/7779/p2p/12D3KooWJQ3k8Gg5vGrUeASnN3Z1E6VGgmq72Gh84G2sNQkXsRcV \
	--p2p-private-key="88fe55b20e47b0e92a02aa59f3f4a1c95cc2ed7b6cb62df04e2a308e3ae274570e9719104c938c881872ced4e34c137a175b279bdb41ece312fc31b5d6e98162" \
	--metrics-port=9093

pathfinder: juno-cached
	./build/juno \
    	--network=sepolia \
    	--log-level=debug \
    	--db-path=./p2p-dbs/node-pathfinder \
    	--p2p \
    	--p2p-peers=/ip4/127.0.0.1/tcp/8888/p2p/12D3KooWF1JrZWQoBiBSjsFSuLbDiDvqcmJQRLaFQLmpVkHA9duk \
    	--p2p-private-key="54a695e2a5d5717d5ba8730efcafe6f17251a1955733cffc55a4085fbf7f5d2c1b4009314092069ef7ca9b364ce3eb3072531c64dfb2799c6bad76720a5bdff0" \
    	--metrics-port=9094
