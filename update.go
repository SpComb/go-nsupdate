package main

import (
	"github.com/miekg/dns"
	"time"
	"fmt"
	"net"
	"log"
)

type Update struct {
	zone	 string
	name	 string
	ttl		 int

	tsig	 map[string]string
	tsigAlgo TSIGAlgorithm
	server	 string
	timeout	 time.Duration
}

func (u *Update) Init(name string, zone string, server string) error {
	if name == "" {
		return fmt.Errorf("Missing name")
	} else {
		u.name = dns.Fqdn(name)
	}

	if zone == "" {
		// guess
		if labels := dns.Split(u.name); len(labels) > 1 {
			u.zone = u.name[labels[1]:]
		} else {
			return fmt.Errorf("Missing zone")
		}
	} else {
		u.zone = dns.Fqdn(zone)
	}

	if server == "" {
		if server, err := discoverZoneServer(u.zone); err != nil {
			return fmt.Errorf("Failed to discver server")
		} else {
			log.Printf("discover server=%v", server)

			u.server = net.JoinHostPort(server, "53")
		}
	} else {
		if _, _, err := net.SplitHostPort(server); err == nil {
			u.server = server
		} else {
			u.server = net.JoinHostPort(server, "53")
		}
	}

	return nil
}

func (u *Update) InitTSIG(name string, secret string, algo TSIGAlgorithm) {
	u.tsig = map[string]string{dns.Fqdn(name): secret}
	u.tsigAlgo = algo
}

func (u *Update) buildAddr(ip net.IP) dns.RR {
	if ip4 := ip.To4(); ip4 != nil {
		return &dns.A{
			Hdr: dns.RR_Header{Name: u.name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: uint32(u.ttl)},
			A:	 ip4,
		}
	}

	if ip6 := ip.To16(); ip6 != nil {
		return &dns.AAAA{
			Hdr:  dns.RR_Header{Name: u.name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: uint32(u.ttl)},
			AAAA: ip6,
		}
	}

	return nil
}
func (u *Update) buildAddrs(addrs *AddrSet) (rs []dns.RR) {
	for _, ip := range addrs.addrs {
		rs = append(rs, u.buildAddr(ip))
	}

	return rs
}

func (u *Update) buildMsg(addrs *AddrSet) *dns.Msg {
	var msg = new(dns.Msg)

	msg.SetUpdate(u.zone)
	msg.RemoveName([]dns.RR{&dns.RR_Header{Name:u.name}})
	msg.Insert(u.buildAddrs(addrs))

	if u.tsig != nil {
		for keyName, _ := range u.tsig {
			msg.SetTsig(keyName, string(u.tsigAlgo), TSIG_FUDGE_SECONDS, time.Now().Unix())
		}
	}

	return msg
}

func (u *Update) query(msg *dns.Msg) (*dns.Msg, error) {
	var client = new(dns.Client)

	client.DialTimeout = u.timeout
	client.ReadTimeout = u.timeout
	client.WriteTimeout = u.timeout

	if u.tsig != nil {
		client.TsigSecret = u.tsig
	}

	msg, _, err := client.Exchange(msg, u.server)

	if err != nil {
		return msg, fmt.Errorf("dns:Client.Exchange ... %v: %v", u.server, err)
	}

	if msg.Rcode == dns.RcodeSuccess {
		return msg, nil
	} else {
		return msg, fmt.Errorf("rcode=%v", dns.RcodeToString[msg.Rcode])
	}
}

func (u *Update) Update(addrs *AddrSet, verbose bool) error {
	q := u.buildMsg(addrs)

	if verbose {
		log.Printf("query:\n%v", q)
	}

	r, err := u.query(q)

	if err != nil {
		return err
	}

	if verbose {
		log.Printf("answer:\n%v", r)
	}

	return nil
}
