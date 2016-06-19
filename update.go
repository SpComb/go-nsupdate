package main

import (
	"github.com/vishvananda/netlink"
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
	tsigAlgo string
	server	 string
	timeout	 time.Duration

	link	netlink.Link
	addrs	map[string]Addr
}

func (u *Update) initTSIG(name string, secret string, algo string) {
	u.tsig = map[string]string{name: secret}
	u.tsigAlgo = algo
}

// Update state for link
func (u *Update) scan(iface string, family int) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("netlink.LinkByName %v: %v", iface, err)
	}

	addrs, err := netlink.AddrList(link, family)
	if err != nil {
		return fmt.Errorf("netlink.AddrList %v: %v", link, err)
	}

	// set
	u.addrs = make(map[string]Addr)

	for _, addr := range addrs {
		u.applyLinkAddr(link, addr)
	}

	return nil
}

func (u *Update) applyLinkAddr(link netlink.Link, addr netlink.Addr) {
	linkUp := link.Attrs().Flags & net.FlagUp != 0

	if addr.Scope >= int(netlink.SCOPE_LINK) {
		return
	}

	u.apply(addr.IP, linkUp)
}

// Update state for address
func (u *Update) apply(ip net.IP, up bool) {
	if up {
		log.Printf("update: up %v", ip)

		u.addrs[ip.String()] = Addr{IP: ip}

	} else {
		log.Printf("update: down %v", ip)

		delete(u.addrs, ip.String())
	}
}

func (u *Update) buildRR() (rs []dns.RR) {
	for _, addr := range u.addrs {
		rs = append(rs, addr.buildRR(u.name, u.ttl))
	}

	return rs
}

func (u *Update) buildMsg() *dns.Msg {
	var msg = new(dns.Msg)

	msg.SetUpdate(u.zone)
	msg.RemoveName([]dns.RR{&dns.RR_Header{Name:u.name}})
	msg.Insert(u.buildRR())

	if u.tsig != nil {
		for keyName, _ := range u.tsig {
			msg.SetTsig(keyName, u.tsigAlgo, TSIG_FUDGE_SECONDS, time.Now().Unix())
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

func (u *Update) update(verbose bool) error {
	q := u.buildMsg()

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
