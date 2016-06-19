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
	TTL			int		`long:"ttl" default:"60"`

	Args		struct {
		Name		string	`description:"DNS Name to update"`
	} `positional-args:"yes"`
}

func main() {
	var options Options

	if _, err := flags.Parse(&options); err != nil {
		log.Fatalf("flags.Parse: %v", err)
		os.Exit(1)
	}

	var update = Update{
		ttl:	 options.TTL,
		timeout: options.Timeout,
	}

	if err := update.Init(options.Args.Name, options.Zone, options.Server); err != nil {
		log.Fatalf("init: %v", err)
	}

	if options.TSIGSecret != "" {
		var name = options.TSIGName

		if name == "" {
			name = options.Args.Name
		}

		log.Printf("using TSIG: %v (algo=%v)", name, options.TSIGAlgorithm)

		update.InitTSIG(name, options.TSIGSecret, options.TSIGAlgorithm)
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
