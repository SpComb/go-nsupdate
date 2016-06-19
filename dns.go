package main

import (
	"github.com/miekg/dns"
	"fmt"
)

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
