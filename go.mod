module github.com/rfyiamcool/raft_kvs

go 1.14

replace github.com/Sirupsen/logrus v1.4.1 => github.com/sirupsen/logrus v1.4.1

require (
	github.com/Sirupsen/logrus v1.4.1
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f
	github.com/dustin/go-humanize v1.0.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/prometheus/client_golang v1.7.1
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2
	go.etcd.io/etcd v3.3.25+incompatible
	go.uber.org/zap v1.15.0
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	google.golang.org/grpc v1.31.1
)
