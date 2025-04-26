rust_uniffi = didcomm-rust/uniffi/
rust_release = $(rust_uniffi)/target/release/
didcomm_rust_lib = didcomm/lib

all: install build test

build: build-rust build-swagger build-go

install: install-air install-swagger

clean: git-clean

build-swagger:
	go get github.com/swaggo/swag && swag init -g ./cmd/api/main.go --parseInternal

build-rust:
	git submodule init && git submodule update 
	cd $(rust_uniffi) && cargo build --release
	# create lib dirctory if it does not exist
	mkdir -p $(didcomm_rust_lib)
	cp $(rust_release)libdidcomm_uniffi.* didcomm/lib

build-go:
	go get ./cmd/api/
	go build ./cmd/api/

install-swagger:
	go install github.com/swaggo/swag/cmd/swag@latest

install-air:
	go install github.com/air-verse/air@latest

dev:
	export LD_LIBRARY_PATH=${PWD}/didcomm/lib && air

test:
	export LD_LIBRARY_PATH=${PWD}/didcomm/lib && go test ./...

git-clean:
	git clean -dfX

docker-build:
	cd deployment/docker && docker build . --tag didcommconnector --build-context files=../..

docker-run:
	docker run -p 9090:9090 -d --name dcc didcommconnector

docker-start:
	docker start dcc

docker-clean:
	docker rm -f dcc
	docker image rm didcommconnector:latest