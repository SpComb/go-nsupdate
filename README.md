# go-nsupdate
Update dynamic DNS records from netlink.

`go-nsupdate` reads interface addresses from netlink, updating on `ip link up/down` and `ip addr add/del` events.

The set of active interface IPv4/IPv6 addresses is used to send DNS `UPDATE` requests to the primary NS for a DNS zone.

## Install

    go get -v github.com/qmsk/go-nsupdate

## Usage

    TSIG_SECRET=... go-nsupdate --interface=vlan-wan --tsig-algorithm=hmac-sha256 yzzrt.dyn.qmsk.net --watch
    
    # Using a generated TSIG key:
    # TSIG_SECRET=$(python -c 'import os; print os.urandom(32).encode("base64")')


