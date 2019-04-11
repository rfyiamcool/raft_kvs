.DEFAULT_GOAL := build-all
.PHONY: boom

build-all: rm build run

export GO15VENDOREXPERIMENT=1

build:
	@ echo "build..."
	go build -o raftexample

rm:
	@ echo "清理..."
	rm -rf raftexample
	rm -rf raftexample-1
	rm -rf raftexample-3
	rm -rf raftexample-2
	rm -rf raftexample-1-snap
	rm -rf raftexample-2-snap
	rm -rf raftexample-3-snap


fmt:
	@ echo "gofmt格式化..."
	go fmt ./

test:
	@ echo "test.."
	go test -v

put:
	@ echo "put.."

	curl -d "key=xiaorui&value=$(shell head -200 /dev/urandom|cksum|cut -d " " -f1)" "http://127.0.0.1:11111/put"
	curl "http://127.0.0.1:11111/get?key=xiaorui"

	curl -d "key=xiaorui&value=$(shell head -200 /dev/urandom|cksum|cut -d " " -f1)" "http://127.0.0.1:22222/put"
	curl "http://127.0.0.1:22222/get?key=xiaorui"

	curl -d "key=xiaorui&value=$(shell head -200 /dev/urandom|cksum|cut -d " " -f1)" "http://127.0.0.1:33333/put"
	curl "http://127.0.0.1:33333/get?key=xiaorui"

batch1w:
	@ echo "batch put.."
	curl -v -d "count=10000" "http://127.0.0.1:11111/batch_put"

batch10w:
	@ echo "batch put.."
	curl -v -d "count=100000" "http://127.0.0.1:11111/batch_put"

batch20w:
	@ echo "batch put.."
	curl -v -d "count=200000&concurrent=5" "http://127.0.0.1:11111/batch_put"

batch50w:
	@ echo "batch put.."
	curl -v -d "count=500000" "http://127.0.0.1:11111/batch_put"

batch100w:
	@ echo "batch put.."
	curl -v -d "count=1000000&concurrent=5" "http://127.0.0.1:11111/batch_put"

check:
	@ echo "check"
	curl "http://127.0.0.1:11111/get?key=10000"
	@ echo "\n\n"
	curl "http://127.0.0.1:22222/get?key=10000"
	@ echo "\n\n"
	curl "http://127.0.0.1:33333/get?key=10000"

run:
	@ echo "run..."
	rm -rf raftexample
	go build -o raftexample
	goreman start
