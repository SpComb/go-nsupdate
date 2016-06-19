package main

import (
	"github.com/vishvananda/netlink"
	"fmt"
)

// zero value is unspec=all
type Family int

func (f *Family) UnmarshalFlag(value string) error {
	switch (value) {
	case "unspec", "all":
		*f = netlink.FAMILY_ALL
	case "inet", "ipv4":
		*f = netlink.FAMILY_V4
	case "inet6", "ipv6":
		*f = netlink.FAMILY_V6
	default:
		return fmt.Errorf("Invalid --family=%v", value)
	}

	return nil
}


