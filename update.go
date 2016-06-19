package main

import (
	"github.com/miekg/dns"
	"time"
	"fmt"
	"net"
	"log"
)

type updateState struct {
	updateZone	string
	removeNames	[]dns.RR
	inserts		[]dns.RR
}

type Update struct {
	ttl		 int
	timeout  time.Duration
	retry	 time.Duration
	verbose	 bool

	zone	 string
	name	 string

	tsig	 map[string]string
	tsigAlgo TSIGAlgorithm
	server	 string

	updateChan chan updateState
	doneChan   chan error
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
func (u *Update) buildState(addrs *AddrSet) (state updateState, err error) {
	state.updateZone = u.zone
	state.removeNames = []dns.RR{
		&dns.RR_Header{Name:u.name},
	}

	addrs.Each(func(ip net.IP){
		state.inserts = append(state.inserts, u.buildAddr(ip))
	})

	return
}
func (u *Update) buildQuery(state updateState) *dns.Msg {
	var msg = new(dns.Msg)

	msg.SetUpdate(state.updateZone)
	msg.RemoveName(state.removeNames)
	msg.Insert(state.inserts)

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

func (u *Update) update(state updateState) error {
	q := u.buildQuery(state)

	if u.verbose {
		log.Printf("update query:\n%v", q)
	} else {
		log.Printf("update query...")
	}

	r, err := u.query(q)

	if err != nil {
		return err
	}

	if u.verbose {
		log.Printf("update answer:\n%v", r)
	} else {
		log.Printf("update answer")
	}

	return nil
}

func (u *Update) run() {
	var state updateState
	var retry = 0
	var retryTimer = time.NewTimer(time.Duration(0))
	var updateChan = u.updateChan
	var updateError error

	defer func(){u.doneChan <-updateError}()

	for {
		select {
		case updateState, running := <-updateChan:
			if running {
				// Update() called
				state = updateState

			} else if retry > 0 {
				// Done() called, but still waiting for retry...
				updateChan = nil
				continue

			} else {
				// Done() called, no retrys or updates remaining
				return
			}

		case <-retryTimer.C:
			if retry == 0 {
				// spurious timer event..
				continue
			}

			// trigger retry
		}

		if err := u.update(state); err != nil {
			log.Printf("update (retry=%v) error: %v", retry, err)

			updateError = err
			retry++
		} else {
			// success
			updateError = nil
			retry = 0
		}

		if retry == 0 && updateChan == nil {
			// done, no more updates
			return

		} else if retry == 0 {
			// wait for next update
			retryTimer.Stop()

		} else {
			retryTimeout := time.Duration(retry * int(u.retry))

			// wait for next retry
			// TODO: exponential backoff?
			retryTimer.Reset(retryTimeout)

			log.Printf("update retry in %v...", retryTimeout)
		}
	}
}

func (u *Update) Start() {
	u.updateChan = make(chan updateState)

	go u.run()
}

func (u *Update) Update(addrs *AddrSet) error {
	if state, err := u.buildState(addrs); err != nil {
		return err
	} else {
		u.updateChan <- state
	}

	return nil
}

func (u *Update) Done() error {
	u.doneChan = make(chan error)

	close(u.updateChan)

	return <-u.doneChan
}
