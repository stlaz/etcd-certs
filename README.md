## prepare a network to run the etcd instances

Docker requires you to specify the subnet CIDR to be able to statically assign IPs to pods, use something like the below.

```bash
$ docker network create -d bridge etcd-network --subnet=172.19.0.0/16
```

## running the scenarios

```bash
$ eval "$(./etcdcerts --network-cidr=172.19.0.0/16 --scenario 1)" ./run-etcd.sh
```

## using a client to query the cluster

After running the eval above, you should already have the `CA_PATH` env variable set.

```bash
$ docker run --net etcd-network --rm -ti -e ETCDCTL_API=3 -v /tmp:/tmp --name etcdclient quay.io/coreos/etcd:v3.3 etcdctl --cert=${CA_PATH}/client/client.crt --key=${CA_PATH}/client/client.key --endpoints=https://${PEER_IP0}:2379 --cacert=/${CA_PATH}/ca.crt get foo
```
