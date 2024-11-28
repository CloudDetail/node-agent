package nettool

import (
	"net"
	"syscall"

	"github.com/Asphaltt/go-iproute2"
	"github.com/mdlayher/netlink"
)

type Tuple struct {
	SrcIp     string
	DstIp     string
	DstPort   uint16
	ServiceIp string
}

// A Client can manipulate ss netlink interface.
type Client struct {
	conn *netlink.Conn
}

// New creates a Client which can issue ss commands.
func New() (*Client, error) {
	conn, err := netlink.Dial(iproute2.FamilySocketMonitoring, nil)
	if err != nil {
		return nil, err
	}

	return NewWithConn(conn), nil
}

func (c *Client) CloseConnection() {
	c.conn.Close()
}

// NewWithConn creates a Client which can issue ss commands using an existing
// netlink connection.
func NewWithConn(conn *netlink.Conn) *Client {
	return &Client{
		conn: conn,
	}
}

func (c *Client) listSockets(req *iproute2.InetDiagReq) ([]*Tuple, error) {
	var msg netlink.Message
	msg.Header.Type = iproute2.MsgTypeSockDiagByFamily
	msg.Header.Flags = netlink.Dump | netlink.Request
	msg.Data, _ = req.MarshalBinary()

	msgs, err := c.conn.Execute(msg)
	if err != nil {
		return nil, err
	}

	entries := make([]*Tuple, 0, len(msgs))
	for _, msg := range msgs {
		if msg.Header.Type != iproute2.MsgTypeSockDiagByFamily {
			continue
		}

		data := msg.Data
		var diagMsg iproute2.InetDiagMsg
		if err := diagMsg.UnmarshalBinary(data); err != nil {
			return entries, err
		}
		if diagMsg.Family != syscall.AF_INET &&
			diagMsg.Family != syscall.AF_INET6 {
			continue
		}

		var e Tuple
		if diagMsg.Family == syscall.AF_INET {
			e.SrcIp = net.IP(diagMsg.Saddr[:4]).String()
			e.DstIp = net.IP(diagMsg.Daddr[:4]).String()
		} else {
			e.SrcIp = net.IP(diagMsg.Saddr[:]).String()
			e.DstIp = net.IP(diagMsg.Daddr[:]).String()
		}
		e.DstPort = diagMsg.Dport

		data = data[iproute2.SizeofInetDiagMsg:]
		ad, err := netlink.NewAttributeDecoder(data)
		if err != nil {
			return entries, err
		}
		for ad.Next() {
			switch ad.Type() {
			}
		}
		if err := ad.Err(); err != nil {
			return entries, err
		}
		//fmt.Println(e)
		entries = append(entries, &e)
	}
	return entries, nil
}

// ListTcp4Conns retrieves all tcp connections from kernel.
func (c *Client) ListTcp4Conns() ([]*Tuple, error) {
	var req iproute2.InetDiagReq
	req.Family = syscall.AF_INET
	req.Protocol = syscall.IPPROTO_TCP
	req.States = uint32(iproute2.Conn)
	return c.listSockets(&req)
}

// ListTcp6Conns retrieves all tcp connections from kernel.
func (c *Client) ListTcp6Conns() ([]*Tuple, error) {
	var req iproute2.InetDiagReq
	req.Family = syscall.AF_INET6
	req.Protocol = syscall.IPPROTO_TCP
	req.States = uint32(iproute2.Conn)
	return c.listSockets(&req)
}
