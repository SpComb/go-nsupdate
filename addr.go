package main

import (
	"net"
	"github.com/vishvananda/netlink"
	"fmt"
	"log"
)

type AddrSet struct {
	addrs	map[string]net.IP
}

func (as *AddrSet) ScanInterface(iface string, family Family) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("netlink.LinkByName %v: %v", iface, err)
	}

	addrList, err := netlink.AddrList(link, int(family))
	if err != nil {
		return fmt.Errorf("netlink.AddrList %v: %v", link, err)
	}

	// set
	as.addrs = make(map[string]net.IP)

	for _, addr := range addrList {
		as.applyLinkAddr(link, addr)
	}

	return nil
}

func (as *AddrSet) applyLinkAddr(link netlink.Link, addr netlink.Addr) {
	linkUp := link.Attrs().Flags & net.FlagUp != 0

	if addr.Scope >= int(netlink.SCOPE_LINK) {
		return
	}

	as.applyAddr(addr.IP, linkUp)
}

// Update state for address
func (as *AddrSet) applyAddr(ip net.IP, up bool) {
	if up {
		log.Printf("update: up %v", ip)

		as.addrs[ip.String()] = ip

	} else {
		log.Printf("update: down %v", ip)

		delete(as.addrs, ip.String())
	}
}
