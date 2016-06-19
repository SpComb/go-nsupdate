package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/vishvananda/netlink"
	"github.com/miekg/dns"
	"net"
	"fmt"
	"log"
	"os"
	"time"
)

// zero value is unspec=all
type Family int

func (f *Family) UnmarshalFlag(value string) error {
	switch (value) {
	case "inet", "ipv4":
		*f = netlink.FAMILY_V4
	case "inet6", "ipv6":
		*f = netlink.FAMILY_V6
	default:
		return fmt.Errorf("Invalid --family=%v", value)
	}

	return nil
}

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

type Options struct {
	Verbose			bool	`long:"verbose" short:"v"`

	Interface		string	`long:"interface" short:"i" value-name:"IFACE" description:"Use address from interface"`
	InterfaceFamily	Family	`long:"interface-family"`

	Server		string	`long:"server" value-name:"HOST[:PORT]"`
	Timeout		time.Duration `long:"timeout" value-name:"DURATION" default:"10s"`
	TSIGName	string	`long:"tsig-name"`
	TSIGSecret	string	`long:"tsig-secret" env:"TSIG_SECRET"`
	TSIGAlgorithm TSIGAlgorithm `long:"tsig-algorithm" default:"hmac-sha1."`

	Zone		string	`long:"zone" description:"Zone to update"`
	Name		string	`long:"name" description:"Name to update"`
	TTL			int		`long:"ttl" default:"60"`
}

func main() {
	var options Options

	if args, err := flags.Parse(&options); err != nil {
		log.Fatalf("flags.Parse: %v", err)
		os.Exit(1)
	} else if len(args) > 0 {
		log.Fatalf("Usage: no args")
		os.Exit(1)
	}

	var update = &Update{
		zone:	 dns.Fqdn(options.Zone),
		name:	 dns.Fqdn(options.Name),
		ttl:	 options.TTL,
		timeout: options.Timeout,
	}

	if _, _, err := net.SplitHostPort(options.Server); err == nil {
		update.server = options.Server
	} else {
		update.server = net.JoinHostPort(options.Server, "53")
	}

	if options.TSIGName != "" {
		log.Printf("using TSIG: %v (algo=%v)", options.TSIGName, options.TSIGAlgorithm)

		update.initTSIG(dns.Fqdn(options.TSIGName), options.TSIGSecret, string(options.TSIGAlgorithm))
	}

	// run
	if options.Interface == "" {

	} else if err := update.scan(options.Interface, int(options.InterfaceFamily)); err != nil {
		log.Fatalf("scan: %v", err)
	}

	if err := update.update(options.Verbose); err != nil {
		log.Fatalf("update: %v", err)
	}
}
