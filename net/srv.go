package net

import "net"

//checkSRV check if SRV needed and get addrs
func CheckSRV(ip *string, port *int) {
	_, addrs, _ := net.LookupSRV("minecraft", "tcp", *ip)
	if len(addrs) != 0 && addrs[0].Target != "" {
		*ip = addrs[0].Target
		*port = int(addrs[0].Port)
	}
}
