package main

import (
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"time"
)

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

	var update = Update{
		ttl:	 options.TTL,
		timeout: options.Timeout,
	}

	if err := update.Init(options.Name, options.Zone, options.Server); err != nil {
		log.Fatalf("init: %v", err)
	}

	if options.TSIGName != "" {
		log.Printf("using TSIG: %v (algo=%v)", options.TSIGName, options.TSIGAlgorithm)

		update.InitTSIG(options.TSIGName, options.TSIGSecret, options.TSIGAlgorithm)
	}

	// addrs
	var addrs = new(AddrSet)

	if options.Interface == "" {

	} else if err := addrs.ScanInterface(options.Interface, options.InterfaceFamily); err != nil {
		log.Fatalf("addrs scan: %v", err)
	}

	// update
	if err := update.Update(addrs, options.Verbose); err != nil {
		log.Fatalf("update: %v", err)
	} else {
		log.Printf("update: ok")
	}
}
