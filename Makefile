SCENARIO ?= 1

build:
	go build -o etcdcerts ./etcdcerts.go

run: build
	eval "$$(./etcdcerts --network-cidr=172.19.0.0/16 --scenario ${SCENARIO})" ./run-etcd.sh

rm:
	docker rm -f etcd1 etcd2 etcd3

clean: rm
	rm -f etcdcerts
