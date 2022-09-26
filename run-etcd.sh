#!/bin/bash

function etcd_args() {
    PEER_PATH="${1}/${2}-peer"
    SERVING_PATH="${1}/${2}-serving"
    
    echo "--initial-advertise-peer-urls=https://${2}:2380\
 --cert-file=${SERVING_PATH}/server.crt\
 --key-file=${SERVING_PATH}/server.key\
 --trusted-ca-file=${1}/ca.crt\
 --client-cert-auth=true\
 --peer-cert-file=${PEER_PATH}/${PEER_FNAME}.crt\
 --peer-key-file=${PEER_PATH}/${PEER_FNAME}.key\
 --peer-trusted-ca-file=${1}/ca.crt\
 --peer-client-cert-auth=true\
 --advertise-client-urls=https://${2}:2379\
 --listen-client-urls=https://0.0.0.0:2379,unixs://${2}:0\
 --listen-peer-urls=https://0.0.0.0:2380\
 --metrics=extensive\
 --listen-metrics-urls=https://0.0.0.0:9978\
 "

}

docker rm -f etcd1 etcd2 etcd3

docker run --net etcd-network --ip $PEER_IP0 -v /tmp:/tmp  --name etcd1 -d quay.io/coreos/etcd:v3.3 etcd $(etcd_args "$CA_PATH" "$PEER_IP0") --initial-cluster default=https://${PEER_IP0}:2380,etcd2=https://${PEER_IP1}:2380,etcd3=https://${PEER_IP2}:2380
docker run --net etcd-network --ip $PEER_IP1 -v /tmp:/tmp  --name etcd2 -d quay.io/coreos/etcd:v3.3 etcd $(etcd_args "$CA_PATH" "$PEER_IP1") --initial-cluster etcd1=https://${PEER_IP0}:2380,default=https://${PEER_IP1}:2380,etcd3=https://${PEER_IP2}:2380
docker run --net etcd-network --ip $PEER_IP2 -v /tmp:/tmp  --name etcd3 -d quay.io/coreos/etcd:v3.3 etcd $(etcd_args "$CA_PATH" "$PEER_IP2") --initial-cluster etcd1=https://${PEER_IP0}:2380,etcd2=https://${PEER_IP1}:2380,default=https://${PEER_IP2}:2380

# --logger=zap \
# --log-level=info \
# --experimental-initial-corrupt-check=true\

# remove the below because resolve works differently ATM
