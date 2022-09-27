package main

import (
	"flag"
	"fmt"
	"net/netip"
	"os"

	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"k8s.io/apiserver/pkg/authentication/user"
)

func main() {
	addressCIDR := flag.String("network-cidr", "", "the cidr of the network to generate certs for")
	scenario := flag.Int("scenario", 1, "1: peer certs are simple client auth certs\n2: peer certs are simple server auth certs")
	flag.Parse()

	if addressCIDR == nil {
		fmt.Fprint(os.Stderr, "no CIDR provided")
		os.Exit(1)
	}

	var err error
	var netPrefix netip.Prefix
	netPrefix, err = netip.ParsePrefix(*addressCIDR)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse the CIDR: %v", err)
		os.Exit(1)
	}

	tmpDir, err := os.MkdirTemp("/tmp", "etcdsigner")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write a new tmp directory: %v", err)
		os.Exit(1)
	}

	var scenarioFn func(dirName string, netPrefix netip.Prefix) error
	switch *scenario {
	case 1:
		scenarioFn = scenario1
	case 2:
		scenarioFn = scenario2
	case 3:
		scenarioFn = scenario3
	case 4:
		scenarioFn = scenario4
	default:
		scenarioFn = scenario1
	}

	if err := scenarioFn(tmpDir, netPrefix); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate certificates: %v", err.Error())
		os.Exit(1)
	}
}

// scenario1:
// peer certs are simple client cert auth certs
// FAILS:
//
// The server with IP 172.19.0.3 is reporting:
// ```
// 2022-09-26 13:28:59.887671 I | embed: rejected connection from "172.19.0.2:40618" (error "remote error: tls: bad certificate", ServerName "")
// 2022-09-26 13:28:59.887783 I | embed: rejected connection from "172.19.0.4:56730" (error "remote error: tls: bad certificate", ServerName "")
// ...
// 2022-09-26 13:55:41.819751 W | rafthttp: health check for peer d7380397c3ec4b90 could not connect: x509: cannot validate certificate for 172.19.0.2 because it doesn't contain any IP SANs (prober "ROUND_TRIPPER_SNAPSHOT")
// 2022-09-26 13:55:41.819781 W | rafthttp: health check for peer d7380397c3ec4b90 could not connect: x509: cannot validate certificate for 172.19.0.2 because it doesn't contain any IP SANs (prober "ROUND_TRIPPER_RAFT_MESSAGE")
// ```
// That means that peer certs are expected to have a SAN
func scenario1(tmpDir string, netPrefix netip.Prefix) error {
	currentAddress := netPrefix.Addr().Next() // the first IP is likely to be taken by docker itself

	s := cryptomaterial.NewCertificateSigner(
		"etcdsigner-1",
		tmpDir,
		10,
	).WithClientCertificates(
		&cryptomaterial.ClientCertificateSigningRequestInfo{
			CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
				Name:         "client",
				ValidityDays: 10,
			},
			UserInfo: &user.DefaultInfo{Name: "etcd-client"},
		},
	)

	fmt.Printf("CA_PATH=%s ", tmpDir)
	fmt.Printf("PEER_FNAME=client ")
	for i := 0; i < 3; i++ {
		currentAddress = currentAddress.Next()
		fmt.Printf("PEER_IP%d=%s ", i, currentAddress.String())
		currentIPString := currentAddress.String()

		if !netPrefix.Contains(currentAddress) {
			return fmt.Errorf("the CIDR provided is too shallow, %v already does not belong", currentIPString)
		}

		s = s.WithClientCertificates(
			&cryptomaterial.ClientCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         fmt.Sprintf("%s-peer", currentIPString),
					ValidityDays: 10,
				},
				UserInfo: &user.DefaultInfo{Name: currentIPString},
			},
		).WithServingCertificates(
			&cryptomaterial.ServingCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         fmt.Sprintf("%s-serving", currentIPString),
					ValidityDays: 10,
				},
				Hostnames: []string{"localhost", "127.0.0.1", currentIPString},
			},
		)
	}
	if _, err := s.Complete(); err != nil {
		return fmt.Errorf("failed to generate certificates: %v", err)
	}

	return nil
}

// scenario2:
// peer certs are simple client cert auth certs
// FAILS:
//
// The server with IP 172.19.0.3 is reporting:
// ```
// 2022-09-26 13:53:31.250172 I | embed: rejected connection from "172.19.0.2:57472" (error "tls: failed to verify client's certificate: x509: certificate specifies an incompatible key usage", ServerName "")
// 2022-09-26 13:53:31.255746 I | embed: rejected connection from "172.19.0.4:53250" (error "tls: failed to verify client's certificate: x509: certificate specifies an incompatible key usage", ServerName "")
// ...
// 2022-09-26 13:57:50.514068 W | rafthttp: health check for peer d7380397c3ec4b90 could not connect: remote error: tls: bad certificate (prober "ROUND_TRIPPER_SNAPSHOT")
// 2022-09-26 13:57:50.514113 W | rafthttp: health check for peer d7380397c3ec4b90 could not connect: remote error: tls: bad certificate (prober "ROUND_TRIPPER_RAFT_MESSAGE")
// ```
// That means that peer certs are expected to have ClientAuth EKU
func scenario2(tmpDir string, netPrefix netip.Prefix) error {
	currentAddress := netPrefix.Addr().Next() // the first IP is likely to be taken by docker itself

	s := cryptomaterial.NewCertificateSigner(
		"etcdsigner-2",
		tmpDir,
		10,
	).WithClientCertificates(
		&cryptomaterial.ClientCertificateSigningRequestInfo{
			CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
				Name:         "client",
				ValidityDays: 10,
			},
			UserInfo: &user.DefaultInfo{Name: "etcd-client"},
		},
	)

	fmt.Printf("CA_PATH=%s ", tmpDir)
	fmt.Printf("PEER_FNAME=server ")
	for i := 0; i < 3; i++ {
		currentAddress = currentAddress.Next()
		fmt.Printf("PEER_IP%d=%s ", i, currentAddress.String())
		currentIPString := currentAddress.String()

		if !netPrefix.Contains(currentAddress) {
			return fmt.Errorf("the CIDR provided is too shallow, %v already does not belong", currentIPString)
		}

		s = s.WithServingCertificates(
			&cryptomaterial.ServingCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         fmt.Sprintf("%s-peer", currentIPString),
					ValidityDays: 10,
				},
				Hostnames: []string{"localhost", "127.0.0.1", currentIPString},
			},
			&cryptomaterial.ServingCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         fmt.Sprintf("%s-serving", currentIPString),
					ValidityDays: 10,
				},
				Hostnames: []string{"localhost", "127.0.0.1", currentIPString},
			},
		)
	}
	if _, err := s.Complete(); err != nil {
		return fmt.Errorf("failed to generate certificates: %v", err)
	}

	return nil
}

// scenario3
// peer certs are server certs with ClientAuth EKU
// PEER CERTS SEEM TO BE FINE, SERVER CERTS SEEM TO BE FAILING:
//
// The server with IP 172.19.0.3 is reporting:
// ```
// 2022-09-27 10:14:00.741166 I | embed: rejected connection from "127.0.0.1:55934" (error "tls: failed to verify client's certificate: x509: certificate specifies an incompatible key usage", ServerName "")
// WARNING: 2022/09/27 10:14:00 Failed to dial 0.0.0.0:2379: connection error: desc = "transport: authentication handshake failed: remote error: tls: bad certificate"; please retry.
// ```
// This means that actually the serving certificate has wrong key usage :-o
func scenario3(tmpDir string, netPrefix netip.Prefix) error {
	currentAddress := netPrefix.Addr().Next() // the first IP is likely to be taken by docker itself

	s := cryptomaterial.NewCertificateSigner(
		"etcdsigner-2",
		tmpDir,
		10,
	).WithClientCertificates(
		&cryptomaterial.ClientCertificateSigningRequestInfo{
			CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
				Name:         "client",
				ValidityDays: 10,
			},
			UserInfo: &user.DefaultInfo{Name: "etcd-client"},
		},
	)

	fmt.Printf("CA_PATH=%s ", tmpDir)
	fmt.Printf("PEER_FNAME=peer ")
	for i := 0; i < 3; i++ {
		currentAddress = currentAddress.Next()
		fmt.Printf("PEER_IP%d=%s ", i, currentAddress.String())
		currentIPString := currentAddress.String()

		if !netPrefix.Contains(currentAddress) {
			return fmt.Errorf("the CIDR provided is too shallow, %v already does not belong", currentIPString)
		}

		s = s.WithPeerCertificiates(
			&cryptomaterial.PeerCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         fmt.Sprintf("%s-peer", currentIPString),
					ValidityDays: 10,
				},
				UserInfo:  &user.DefaultInfo{Name: "etcd-peer"},
				Hostnames: []string{"localhost", "127.0.0.1", currentIPString},
			},
		).WithServingCertificates(
			&cryptomaterial.ServingCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         fmt.Sprintf("%s-serving", currentIPString),
					ValidityDays: 10,
				},
				Hostnames: []string{"localhost", "127.0.0.1", currentIPString},
			},
		)
	}
	if _, err := s.Complete(); err != nil {
		return fmt.Errorf("failed to generate certificates: %v", err)
	}

	return nil
}

// scenario4
// The serving certs have ClientAuth EKU.
// This finally seems to work.
func scenario4(tmpDir string, netPrefix netip.Prefix) error {
	currentAddress := netPrefix.Addr().Next() // the first IP is likely to be taken by docker itself

	s := cryptomaterial.NewCertificateSigner(
		"etcdsigner-2",
		tmpDir,
		10,
	).WithClientCertificates(
		&cryptomaterial.ClientCertificateSigningRequestInfo{
			CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
				Name:         "client",
				ValidityDays: 10,
			},
			UserInfo: &user.DefaultInfo{Name: "etcd-client"},
		},
	)

	fmt.Printf("CA_PATH=%s ", tmpDir)
	fmt.Printf("PEER_FNAME=peer ")
	fmt.Printf("SERVING_FNAME=peer ")
	for i := 0; i < 3; i++ {
		currentAddress = currentAddress.Next()
		fmt.Printf("PEER_IP%d=%s ", i, currentAddress.String())
		currentIPString := currentAddress.String()

		if !netPrefix.Contains(currentAddress) {
			return fmt.Errorf("the CIDR provided is too shallow, %v already does not belong", currentIPString)
		}

		s = s.WithPeerCertificiates(
			&cryptomaterial.PeerCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         fmt.Sprintf("%s-peer", currentIPString),
					ValidityDays: 10,
				},
				UserInfo:  &user.DefaultInfo{Name: "etcd-peer"},
				Hostnames: []string{"localhost", "127.0.0.1", currentIPString},
			},
		).WithPeerCertificiates(
			&cryptomaterial.PeerCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         fmt.Sprintf("%s-serving", currentIPString),
					ValidityDays: 10,
				},
				UserInfo:  &user.DefaultInfo{Name: "etcd-serving"},
				Hostnames: []string{"localhost", "127.0.0.1", currentIPString},
			},
		)
	}
	if _, err := s.Complete(); err != nil {
		return fmt.Errorf("failed to generate certificates: %v", err)
	}

	return nil
}
