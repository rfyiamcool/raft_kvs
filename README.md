# raft_kvs

* benchmark raft replication log performance
* test how long does follower switch leader

`some code refer from go.etcd.io/etcd/contrib/raftexample`

## Modify

* add isLeader api
* add getSelfState api
* more code description
* fix don't rm old snapshot bug
* fix don't rm old wals bug
* fix restore data raise block bug in snapshot 
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

make batch100w
make check
```

## Method

**show leader and followrer**

```
make leader
```

**check a kv record, check value in all nodes**

```
make check
```

**show raft nodes info**

```
make info
```

## Benchmark

benchmark raft log replication performance

#### benchmark method

* In a leader node, active make 100w record, and sync data to followers. 
* leader and followers all in a node.

`8 cpu core, 16 mem`

#### Result

**Concurrent 1 workers**

```
7 w QPS/S
```

**Concurrent 20 workers**

```
15 w QPS/S
```

**Concurrent 100 workers**

```
17 w QPS/S
```

**Concurrent 500 workers**

```
14 w QPS/S
```