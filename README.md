# raft_kvs

* benchmark raft replication log performance
* test how long does follower switch leader

`refer code from go.etcd.io/etcd/contrib/raftexample`

## Modify

* more code description
* format code
* support leader judge
* modify http api
* add benchmark
* more...

## Test

**config dep**

```sh
export GOPATH=<directory>
cd <directory>/src/go.etcd.io/etcd/contrib/raftexample
go build -o raftexample
```

**install goreman**

```
go get github.com/mattn/goreman
```

**run raft cluster**

```
make run
```

**run benchmark**

```
make batch1w
make check

make batch10w
make check

make batch20w
make check
```
