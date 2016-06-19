package main

import (
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"time"
)

type Options struct {
	Verbose			bool	`long:"verbose" short:"v"`
	Watch			bool	`long:"watch" description:"Watch for interface changes"`

	Interface		string	`long:"interface" short:"i" value-name:"IFACE" description:"Use address from interface"`
	InterfaceFamily	Family	`long:"interface-family"`

	Server		string	`long:"server" value-name:"HOST[:PORT]"`
	Timeout		time.Duration `long:"timeout" value-name:"DURATION" default:"10s"`
	Retry		time.Duration `long:"retry" value-name:"DURATION" default:"30s"`
	TSIGName	string	`long:"tsig-name"`
	TSIGSecret	string	`long:"tsig-secret" env:"TSIG_SECRET"`
	TSIGAlgorithm TSIGAlgorithm `long:"tsig-algorithm" default:"hmac-sha1."`

	Zone		string	`long:"zone" description:"Zone to update"`
	TTL			time.Duration `long:"ttl" default:"60s"`

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
		ttl:	 int(options.TTL.Seconds()),
		timeout: options.Timeout,
		retry:   options.Retry,
		verbose: options.Verbose,
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
	addrs, err := InterfaceAddrs(options.Interface, options.InterfaceFamily)
	if err != nil {
		log.Fatalf("addrs scan: %v", err)
	}

	// update
	update.Start()

	for {
		log.Printf("update...")

		if err := update.Update(addrs); err != nil {
			log.Fatalf("update: %v", err)
		}

		if !options.Watch {
			break
		}

		if err := addrs.Read(); err != nil {
			log.Fatalf("addrs read: %v", err)
		} else {
			log.Printf("addrs update...")
		}
	}

	log.Printf("wait...")

	if err := update.Done(); err != nil {
		log.Printf("update done: %v", err)
	} else {
		log.Printf("update done")
	}
}
