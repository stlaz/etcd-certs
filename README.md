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

## Scenarios description

1. Peer certificate is just a client certificate
2. Peer certificate is just a serving certificate
3. Peer certificate is a serving certificate with a random CN and ClientAuth EKU
4. Like 3 but with the **serving** certitifacte having ClientAuth EKU

### Scenario 1

Peer certs are simple client cert auth certs

**FAILS:**

The server with IP 172.19.0.3 is reporting:
```
2022-09-26 13:28:59.887671 I | embed: rejected connection from "172.19.0.2:40618" (error "remote error: tls: bad certificate", ServerName "")
2022-09-26 13:28:59.887783 I | embed: rejected connection from "172.19.0.4:56730" (error "remote error: tls: bad certificate", ServerName "")
...
2022-09-26 13:55:41.819751 W | rafthttp: health check for peer d7380397c3ec4b90 could not connect: x509: cannot validate certificate for 172.19.0.2 because it doesn't contain any IP SANs (prober "ROUND_TRIPPER_SNAPSHOT")
2022-09-26 13:55:41.819781 W | rafthttp: health check for peer d7380397c3ec4b90 could not connect: x509: cannot validate certificate for 172.19.0.2 because it doesn't contain any IP SANs (prober "ROUND_TRIPPER_RAFT_MESSAGE")
```

That means that peer certs **are expected to have a SAN**

### Scenario 2

Reer certs are simple serving certs

**FAILS:**

The server with IP 172.19.0.3 is reporting:
```
2022-09-26 13:53:31.250172 I | embed: rejected connection from "172.19.0.2:57472" (error "tls: failed to verify client's certificate: x509: certificate specifies an incompatible key usage", ServerName "")
2022-09-26 13:53:31.255746 I | embed: rejected connection from "172.19.0.4:53250" (error "tls: failed to verify client's certificate: x509: certificate specifies an incompatible key usage", ServerName "")
...
2022-09-26 13:57:50.514068 W | rafthttp: health check for peer d7380397c3ec4b90 could not connect: remote error: tls: bad certificate (prober "ROUND_TRIPPER_SNAPSHOT")
2022-09-26 13:57:50.514113 W | rafthttp: health check for peer d7380397c3ec4b90 could not connect: remote error: tls: bad certificate (prober "ROUND_TRIPPER_RAFT_MESSAGE")
```
That means that peer certs **are expected to have ClientAuth EKU**.

### Scenario 3

Peer certs are server certs with ClientAuth EKU

**PEER CERTS SEEM TO BE FINE, SERVER CERTS SEEM TO BE FAILING:**

The server with IP 172.19.0.3 is reporting:
```
2022-09-27 10:14:00.741166 I | embed: rejected connection from "127.0.0.1:55934" (error "tls: failed to verify client's certificate: x509: certificate specifies an incompatible key usage", ServerName "")
WARNING: 2022/09/27 10:14:00 Failed to dial 0.0.0.0:2379: connection error: desc = "transport: authentication handshake failed: remote error: tls: bad certificate"; please retry.
```
This means that actually the **serving certificate is expected to have ClientAuth EKU**.

### Scenario 4

The serving certs have ClientAuth EKU.

**This finally seems to work.**
