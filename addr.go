package main

import (
	"net"
	"github.com/miekg/dns"
)

type Addr struct {
	IP		net.IP
}

func (addr Addr) buildRR(name string, ttl int) dns.RR {
	if ip4 := addr.IP.To4(); ip4 != nil {
		return &dns.A{
			Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: uint32(ttl)},
			A:	 ip4,
		}
	}

	if ip6 := addr.IP.To16(); ip6 != nil {
		return &dns.AAAA{
			Hdr:  dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: uint32(ttl)},
			AAAA: ip6,
		}
	}

	return nil
}

