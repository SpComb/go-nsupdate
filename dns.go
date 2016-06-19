package main

import (
	"github.com/miekg/dns"
	"fmt"
	"time"
	"log"
	"net"
)

const TSIG_FUDGE_SECONDS = 300
type TSIGAlgorithm string

func (t *TSIGAlgorithm) UnmarshalFlag(value string) error {
	switch (value) {
	case "hmac-md5", "md5":
		*t = dns.HmacMD5
	case "hmac-sha1", "sha1":
		*t = dns.HmacSHA1
	case "hmac-sha256", "sha256":
		*t = dns.HmacSHA256
	case "hmac-sha512", "sha512":
		*t = dns.HmacSHA512
	default:
		return fmt.Errorf("Invalid --tsig-algorithm=%v", value)
	}

	return nil
}

func query(query *dns.Msg) (*dns.Msg, error) {
	clientConfig, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(clientConfig.Timeout * int(time.Second))

	var client = new(dns.Client)

	client.DialTimeout = timeout
	client.ReadTimeout = timeout
	client.WriteTimeout = timeout

	for _, server := range clientConfig.Servers {
		addr := net.JoinHostPort(server, "53")

		if answer, _, err := client.Exchange(query, addr); err != nil {
			log.Printf("query %v: %v", server, err)
			continue
		} else {
			return answer, nil
		}
	}

	return nil, fmt.Errorf("DNS query failed")
}

// Discover likely master NS for zone
func discoverZoneServer(zone string) (string, error) {
	var q = new(dns.Msg)

	q.SetQuestion(zone, dns.TypeSOA)

	r, err := query(q)
	if err != nil {
		return "", err
	}

	for _, rr := range r.Answer {
		if soa, ok := rr.(*dns.SOA); ok {
			return soa.Ns, nil
		}
	}

	return "", fmt.Errorf("No SOA response")
}
