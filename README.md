# go-nsupdate
Update dynamic DNS records from netlink.

`go-nsupdate` reads interface addresses from netlink, updating on `ip link up/down` and `ip addr add/del` events.

The set of active interface IPv4/IPv6 addresses is used to send DNS `UPDATE` requests to the primary NS for a DNS zone.

The DNS update requests are retried in the background (XXX: currently blocks for 10s on each query attempt).

## Install

    go get -v github.com/qmsk/go-nsupdate

## Usage

    
    # Using a generated TSIG key:
    # TSIG_SECRET=$(python -c 'import os; print os.urandom(32).encode("base64")')
    
    TSIG_SECRET=... go-nsupdate --interface=vlan-wan --tsig-algorithm=hmac-sha256 yzzrt.dyn.qmsk.net --watch
    2016/06/19 21:29:33 discover server=zovoweix.qmsk.net.
    2016/06/19 21:29:33 using TSIG: yzzrt.dyn.qmsk.net (algo=hmac-sha256.)
    2016/06/19 21:29:33 AddrSet iface=vlan-wan: up 2001:14ba:400:0:7:1449:a833:f11f
    2016/06/19 21:29:33 update...
    2016/06/19 21:29:33 update query:
    ;; opcode: UPDATE, status: NOERROR, id: 61616
    ;; flags:; QUERY: 1, ANSWER: 0, AUTHORITY: 2, ADDITIONAL: 1

    ;; QUESTION SECTION:
    ;dyn.qmsk.net.  IN       SOA

    ;; AUTHORITY SECTION:
    yzzrt.dyn.qmsk.net.     0       ANY     ANY
    yzzrt.dyn.qmsk.net.     60      IN      AAAA    2001:14ba:400:0:7:1449:a833:f11f

    ;; ADDITIONAL SECTION:

    ;; TSIG PSEUDOSECTION:
    yzzrt.dyn.qmsk.net.     0       ANY     TSIG     hmac-sha256. 20160619182933 300 0  61616 0 0 
    2016/06/19 21:29:33 update answer:
    ;; opcode: UPDATE, status: NOERROR, id: 61616
    ;; flags: qr; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 1

    ;; QUESTION SECTION:
    ;dyn.qmsk.net.  IN       SOA

    ;; ADDITIONAL SECTION:

    ;; TSIG PSEUDOSECTION:
    yzzrt.dyn.qmsk.net.     0       ANY     TSIG     hmac-sha256. 20160619182933 300 32 E083433B2893B2036B24549E3537C6E17B858019B9862DC2EB9EDFB959D03232 61616 0 0 
    2016/06/19 21:46:34 AddrSet iface=vlan-wan: up 213.243.178.191
    2016/06/19 21:46:34 addrs update...
    2016/06/19 21:46:34 update...
    2016/06/19 21:46:34 update query:
    ;; opcode: UPDATE, status: NOERROR, id: 30973
    ;; flags:; QUERY: 1, ANSWER: 0, AUTHORITY: 3, ADDITIONAL: 1

    ;; QUESTION SECTION:
    ;dyn.qmsk.net.  IN       SOA

    ;; AUTHORITY SECTION:
    yzzrt.dyn.qmsk.net.     0       ANY     ANY
    yzzrt.dyn.qmsk.net.     60      IN      AAAA    2001:14ba:400:0:7:1449:a833:f11f
    yzzrt.dyn.qmsk.net.     60      IN      A       213.243.178.191

    ;; ADDITIONAL SECTION:

    ;; TSIG PSEUDOSECTION:
    yzzrt.dyn.qmsk.net.     0       ANY     TSIG     hmac-sha256. 20160619184634 300 0  30973 0 0 
    2016/06/19 21:46:35 update answer:
    ;; opcode: UPDATE, status: NOERROR, id: 30973
    ;; flags: qr; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 1

    ;; QUESTION SECTION:
    ;dyn.qmsk.net.  IN       SOA

    ;; ADDITIONAL SECTION:

    ;; TSIG PSEUDOSECTION:
    yzzrt.dyn.qmsk.net.     0       ANY     TSIG     hmac-sha256. 20160619184635 300 32 1F7F1EB8A3D5213EAAA163AE78388D48911495A0F3E2870688F3338160905EC9 30973 0

