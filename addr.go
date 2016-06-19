package main

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"io"
	"log"
	"net"
)

type AddrSet struct {
	linkAttrs netlink.LinkAttrs
	linkChan  chan netlink.LinkUpdate
	addrChan  chan netlink.AddrUpdate

	addrs map[string]net.IP
}

func (addrs *AddrSet) String() string {
	return fmt.Sprintf("AddrSet iface=%v", addrs.linkAttrs.Name)
}

func (addrs *AddrSet) testFlag(flag net.Flags) bool {
	return addrs.linkAttrs.Flags&flag != 0
}

func (addrs *AddrSet) Up() bool {
	return addrs.testFlag(net.FlagUp)
}

func InterfaceAddrs(iface string, family Family) (*AddrSet, error) {
	var addrs AddrSet

	link, err := netlink.LinkByName(iface)
	if err != nil {
		return nil, fmt.Errorf("netlink.LinkByName %v: %v", iface, err)
	} else {
		addrs.linkAttrs = *link.Attrs()
	}

	// list
	if addrList, err := netlink.AddrList(link, int(family)); err != nil {
		return nil, fmt.Errorf("netlink.AddrList %v: %v", link, err)
	} else {
		addrs.addrs = make(map[string]net.IP)

		for _, addr := range addrList {
			addrs.updateAddr(addr, true)
		}
	}

	// update
	addrs.linkChan = make(chan netlink.LinkUpdate)
	addrs.addrChan = make(chan netlink.AddrUpdate)

	if err := netlink.LinkSubscribe(addrs.linkChan, nil); err != nil {
		return nil, fmt.Errorf("netlink.LinkSubscribe: %v", err)
	}

	if err := netlink.AddrSubscribe(addrs.addrChan, nil); err != nil {
		return nil, fmt.Errorf("netlink.AddrSubscribe: %v", err)
	}

	return &addrs, nil
}

func (addrs *AddrSet) Read() error {
	for {
		select {
		case linkUpdate, ok := <-addrs.linkChan:
			if !ok {
				return io.EOF
			}

			linkAttrs := linkUpdate.Attrs()

			if linkAttrs.Index != addrs.linkAttrs.Index {
				continue
			}

			// update state
			addrs.updateLink(*linkAttrs)

		case addrUpdate, ok := <-addrs.addrChan:
			if !ok {
				return io.EOF
			}

			if addrUpdate.LinkIndex != addrs.linkAttrs.Index {
				continue
			}

			// XXX: scope and other filters?
			addrs.updateAddr(addrUpdate.Addr, addrUpdate.NewAddr)

			return nil
		}
	}
}

// Update state for address
func (addrs *AddrSet) updateAddr(addr netlink.Addr, up bool) {
	if addr.Scope >= int(netlink.SCOPE_LINK) {
		return
	}

	ip := addr.IP

	if up {
		log.Printf("%v: up %v", addrs, ip)

		addrs.addrs[ip.String()] = ip

	} else {
		log.Printf("%v: down %v", addrs, ip)

		delete(addrs.addrs, ip.String())
	}
}

func (addrs *AddrSet) updateLink(linkAttrs netlink.LinkAttrs) {
	addrs.linkAttrs = linkAttrs

	if !addrs.Up() {
		log.Printf("%v: down", addrs)
	}
}

func (addrs *AddrSet) Each(visitFunc func(net.IP)) {
	if !addrs.Up() {
		// link down has no up addrs
		return
	}

	for _, ip := range addrs.addrs {
		visitFunc(ip)
	}
}
