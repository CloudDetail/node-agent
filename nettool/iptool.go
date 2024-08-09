package nettool

import "net"

func IPv4MappedToIPv6(ipv6Address string) (bool, string) {
	addr, err := net.ResolveTCPAddr("tcp", ipv6Address)
	if err != nil {
		return false, "" // 解析出错，不是IPv4-mapped
	}

	ipv4Address := addr.IP.To4()
	ipString := ""
	if ipv4Address != nil {
		ipString = ipv4Address.String()
	}
	return ipv4Address != nil, ipString
}
