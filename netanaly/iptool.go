package netanaly

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"pinger/nettool"
	"strconv"
	"strings"
	"sync"
)

func IPv4MappedToIPv6(ipv6Address string) (bool, string) {
	if !strings.HasPrefix(ipv6Address, "::ffff:") {
		return false, ""
	}

	// 提取IPv4地址部分
	ipv4Part := ipv6Address[len("::ffff:"):]
	ip, _, err := net.SplitHostPort(ipv4Part)
	stringIp := ""
	if err == nil {
		stringIp = ip
	}

	return true, stringIp
}

func hexToByte(s byte) (byte, error) {
	if s <= '9' && s >= '0' {
		return s - '0', nil
	} else if s >= 'a' && s <= 'f' {
		return s - 'a' + 10, nil
	} else if s >= 'A' && s <= 'F' {
		return s - 'A' + 10, nil
	} else {
		return s, errors.New("invalid byte")
	}
}

func AnalysisIpv6(s string) (string, error) {
	hexIP := s[:len(s)-5]
	hexPort := s[len(s)-4:]
	if hexIP == "00000000000000000000000000000000" {
		// 忽略0.0.0.0
		return "", nil
	}
	b := make(net.IP, 16)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			srcIndex := i*4 + 3 - j
			targetIndex := i*4 + j
			c, err := hexToByte(hexIP[srcIndex*2])
			if err != nil {
				return "", err
			}
			d, err := hexToByte(hexIP[srcIndex*2+1])
			if err != nil {
				return "", err
			}
			b[targetIndex] = c*16 + d
		}
	}
	ip := b.String()
	l := strings.Split(ip, ".")
	if len(l) > 2 {
		// 判断是否是真的ipv6格式， 如果不是则转为ipv6格式
		ip = "::ffff:" + ip
	}
	port, err := strconv.ParseUint(hexPort, 16, 16)
	return fmt.Sprintf("%s:%d", ip, port), err
}

// AnalysisIpv4 16进制的ipv4地址转为可识别的ipv4格式：例如“10.10.25.50:8888”
func AnalysisIpv4(s string) (string, error) {
	hexIP := s[:len(s)-5]
	hexPort := s[len(s)-4:]
	bytesIP, err := hex.DecodeString(hexIP)
	if err != nil {
		return "", nil
	}
	uint32IP := binary.LittleEndian.Uint32(bytesIP) //转换为主机字节序
	IP := make(net.IP, 4)
	binary.BigEndian.PutUint32(IP, uint32IP)
	port, err := strconv.ParseUint(hexPort, 16, 16)
	return fmt.Sprintf("%s:%d", IP.String(), port), err
}

var stateMap = map[string]string{
	"01": "ESTABLISHED",
	"02": "SYN_SENT",
	"03": "SYN_RECV",
	"04": "FIN_WAIT1",
	"05": "FIN_WAIT2",
	"06": "TIME_WAIT",
	"07": "CLOSE",
	"08": "CLOSE_WAIT",
	"09": "LAST_ACK",
	"0A": "LISTEN",
	"0B": "CLOSING",
}

type RttStatistic struct {
	SumLatency float64
	Count      int
	Pids       []uint32
}

var GlobalRttResultMap = make(map[nettool.Tuple]RttStatistic)
var RttResultMapMutex = &sync.Mutex{}
