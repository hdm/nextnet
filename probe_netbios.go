package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

const MaxPendingReplies int = 256
const MaxProbeResponseTime time.Duration = time.Second * 2

type NetbiosInfo struct {
	statusRecv  time.Time
	nameSent    time.Time
	nameRecv    time.Time
	statusReply NetbiosReplyStatus
	nameReply   NetbiosReplyStatus
}

type ProbeNetbios struct {
	Probe
	socket  net.PacketConn
	replies map[string]*NetbiosInfo
}

type NetbiosReplyHeader struct {
	XID             uint16
	Flags           uint16
	QuestionCount   uint16
	AnswerCount     uint16
	AuthCount       uint16
	AdditionalCount uint16
	QuestionName    [34]byte
	RecordType      uint16
	RecordClass     uint16
	RecordTTL       uint32
	RecordLength    uint16
}

type NetbiosReplyName struct {
	Name [15]byte
	Type uint8
	Flag uint16
}

type NetbiosReplyAddress struct {
	Flag    uint16
	Address [4]uint8
}

type NetbiosReplyStatus struct {
	Header    NetbiosReplyHeader
	HostName  [15]byte
	UserName  [15]byte
	Names     []NetbiosReplyName
	Addresses []NetbiosReplyAddress
	HWAddr    string
}

func (p *ProbeNetbios) ProcessReplies() {
	defer p.waiter.Done()

	buff := make([]byte, 1500)

	p.replies = make(map[string]*NetbiosInfo)

	for {
		rlen, raddr, err := p.socket.ReadFrom(buff)
		if err != nil {
			switch err := err.(type) {

			case *net.OpError:
				log.Println("OpError:", err)
				return

			case net.Error:
				if err.Timeout() {
					log.Printf("probe %s receiver timed out: %s", p, err)
					return
				}

				log.Printf("probe %s receiver returned net error: %s", p, err)
				return

			default:
				log.Printf("probe %s receiver returned error: %s", p, err)
				return
			}
		}

		if raddr == nil {
			log.Printf("reply address is nil, skipping...")
			return
		}

		reply, err := p.ParseReply(buff[0 : rlen-1])
		if err != nil {
			log.Println("parse reply failed:", err)
			continue
		}

		if len(reply.Names) == 0 && len(reply.Addresses) == 0 {
			continue
		}

		ip := raddr.(*net.UDPAddr).IP.String()
		if _, ok := p.replies[ip]; !ok {
			p.replies[ip] = new(NetbiosInfo)
		}

		// Handle status replies by sending a name request
		if reply.Header.RecordType == 0x21 {
			// log.Printf("probe %s received a status reply of %d bytes from %s", p, rlen, raddr)
			p.replies[ip].statusReply = reply
			p.replies[ip].statusRecv = time.Now()

			ntime := time.Time{}
			if p.replies[ip].nameSent == ntime {
				p.replies[ip].nameSent = time.Now()
				p.SendNameRequest(ip)
			}
		}

		// Handle name replies by reporting the result
		if reply.Header.RecordType == 0x20 {
			// log.Printf("probe %s received a name reply of %d bytes from %s", p, rlen, raddr)
			p.replies[ip].nameReply = reply
			p.replies[ip].nameRecv = time.Now()
			p.ReportResult(ip)
		}
	}
}

func (p *ProbeNetbios) SendRequest(ip string, req []byte) {
	addr, aerr := net.ResolveUDPAddr("udp", ip+":137")
	if aerr != nil {
		log.Printf("probe %s failed to resolve %s (%s)", p, ip, aerr)
		return
	}

	// Retry in case of network buffer congestion
	wcnt := 0
	for wcnt = 0; wcnt < 5; wcnt++ {

		p.CheckRateLimit()

		_, werr := p.socket.WriteTo(req, addr)
		if werr != nil {
			log.Printf("probe %s [%d/%d] failed to send to %s (%s)", p, wcnt+1, 5, ip, werr)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	// Were we able to send it eventually?
	if wcnt == 5 {
		log.Printf("probe %s [%d/%d] gave up sending to %s", p, wcnt, 5, ip)
	}
}

func (p *ProbeNetbios) SendStatusRequest(ip string) {
	// log.Printf("probe %s is sending a status request to %s", p, ip)
	p.SendRequest(ip, p.CreateStatusRequest())
}

func (p *ProbeNetbios) SendNameRequest(ip string) {
	sreply := p.replies[ip].statusReply
	name := TrimName(string(sreply.HostName[:]))
	p.SendRequest(ip, p.CreateNameRequest(name))
}

func (p *ProbeNetbios) ResultFromIP(ip string) scanResult {
	sreply := p.replies[ip].statusReply
	nreply := p.replies[ip].nameReply

	res := scanResult{
		Host:  ip,
		Port:  "137",
		Proto: "udp",
		Probe: p.String(),
	}

	res.Info = make(map[string]string)

	res.Name = TrimName(string(sreply.HostName[:]))

	if nreply.Header.RecordType == 0x20 {
		for _, ainfo := range nreply.Addresses {

			net := fmt.Sprintf("%d.%d.%d.%d", ainfo.Address[0], ainfo.Address[1], ainfo.Address[2], ainfo.Address[3])
			if net == "0.0.0.0" {
				continue
			}

			res.Nets = append(res.Nets, net)
		}
	}

	if sreply.HWAddr != "00:00:00:00:00:00" {
		res.Info["hwaddr"] = sreply.HWAddr
	}

	username := TrimName(string(sreply.UserName[:]))
	if len(username) > 0 && username != res.Name {
		res.Info["username"] = username
	}

	for _, rname := range sreply.Names {

		tname := TrimName(string(rname.Name[:]))
		if tname == res.Name {
			continue
		}

		if rname.Flag&0x0800 != 0 {
			continue
		}

		res.Info["domain"] = tname
	}

	return res
}

func (p *ProbeNetbios) ReportResult(ip string) {
	p.output <- p.ResultFromIP(ip)
	delete(p.replies, ip)
}

func (p *ProbeNetbios) ReportIncompleteResults() {
	for ip := range p.replies {
		p.ReportResult(ip)
	}
}

func (p *ProbeNetbios) EncodeNetbiosName(name [16]byte) [32]byte {
	encoded := [32]byte{}

	for i := 0; i < 16; i++ {
		if name[i] == 0 {
			encoded[(i*2)+0] = 'C'
			encoded[(i*2)+1] = 'A'
		} else {
			encoded[(i*2)+0] = byte((name[i] / 16) + 0x41)
			encoded[(i*2)+1] = byte((name[i] % 16) + 0x41)
		}
	}

	return encoded
}

func (p *ProbeNetbios) DecodeNetbiosName(name [32]byte) [16]byte {
	decoded := [16]byte{}

	for i := 0; i < 16; i++ {
		if name[(i*2)+0] == 'C' && name[(i*2)+1] == 'A' {
			decoded[i] = 0
		} else {
			decoded[i] = ((name[(i*2)+0] * 16) - 0x41) + (name[(i*2)+1] - 0x41)
		}
	}
	return decoded
}

func (p *ProbeNetbios) ParseReply(buff []byte) (NetbiosReplyStatus, error) {
	resp := NetbiosReplyStatus{}
	temp := bytes.NewBuffer(buff)

	if err := binary.Read(temp, binary.BigEndian, &resp.Header); err != nil && err != io.EOF {
		return resp, fmt.Errorf("failed to read netbios reply status: %s\n", err)
	}

	if resp.Header.QuestionCount != 0 {
		return resp, errors.New("question count is not 0")
	}

	if resp.Header.AnswerCount == 0 {
		return resp, errors.New("answer count is 0")
	}

	// Names
	if resp.Header.RecordType == 0x21 {
		var rcnt uint8
		var ridx uint8
		if err := binary.Read(temp, binary.BigEndian, &rcnt); err != nil {
			return resp, err
		}

		for ridx = 0; ridx < rcnt; ridx++ {
			name := NetbiosReplyName{}
			if err := binary.Read(temp, binary.BigEndian, &name); err != nil {
				log.Println("failed to read netbios reply name:", err)
				continue
			}
			resp.Names = append(resp.Names, name)

			if name.Type == 0x20 {
				resp.HostName = name.Name
			}

			if name.Type == 0x03 {
				resp.UserName = name.Name
			}
		}

		var hwbytes [6]uint8
		if err := binary.Read(temp, binary.BigEndian, &hwbytes); err != nil {
			return resp, fmt.Errorf("failed to read hwaddr: %s", err)
		}
		resp.HWAddr = fmt.Sprintf("%.2x:%.2x:%.2x:%.2x:%.2x:%.2x",
			hwbytes[0], hwbytes[1], hwbytes[2], hwbytes[3], hwbytes[4], hwbytes[5],
		)
		return resp, nil
	}

	// Addresses
	if resp.Header.RecordType == 0x20 {
		var ridx uint16
		for ridx = 0; ridx < (resp.Header.RecordLength / 6); ridx++ {
			addr := NetbiosReplyAddress{}
			if err := binary.Read(temp, binary.BigEndian, &addr); err != nil && err != io.EOF {
				log.Println("failed to read netbios reply address:", err)
				continue
			}
			resp.Addresses = append(resp.Addresses, addr)
		}
	}

	return resp, nil
}

func (p *ProbeNetbios) CreateStatusRequest() []byte {
	return []byte{
		byte(rand.Intn(256)), byte(rand.Intn(256)),
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x20, 0x43, 0x4b, 0x41, 0x41, 0x41,
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x41, 0x41, 0x41, 0x00, 0x00, 0x21, 0x00, 0x01,
	}
}

func (p *ProbeNetbios) CreateNameRequest(name string) []byte {
	nbytes := [16]byte{}
	copy(nbytes[0:15], []byte(strings.ToUpper(name)[:]))

	req := []byte{
		byte(rand.Intn(256)), byte(rand.Intn(256)),
		0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x20,
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x00, 0x00, 0x20, 0x00, 0x01,
	}

	encoded := p.EncodeNetbiosName(nbytes)
	copy(req[13:45], encoded[0:32])
	return req
}

func (p *ProbeNetbios) Initialize() {
	p.Setup()
	p.name = "netbios"
	p.waiter.Add(2)

	// Open socket
	var err error
	p.socket, err = net.ListenPacket("udp", "")
	if err != nil {
		log.Println("Listen UDP packet error:", err)
	}

	go p.ProcessReplies()
	go func() {

		for dip := range p.input {
			p.SendStatusRequest(dip)

			// If our pending replies gets > MAX, stop, process, report, clear, resume
			if len(p.replies) > MaxPendingReplies {
				log.Printf("probe %s is flushing due to maximum replies hit (%d)", p, len(p.replies))
				time.Sleep(MaxProbeResponseTime)
				p.ReportIncompleteResults()
			}
		}

		// Sleep for packet timeout of initial probe
		log.Printf("probe %s is waiting for final replies to status probe", p)
		time.Sleep(MaxProbeResponseTime)

		// The receiver is sending interface probes in response to status probes

		log.Printf("probe %s is waiting for final replies to interface probe", p)
		time.Sleep(MaxProbeResponseTime)

		// Shut down receiver
		if err := p.socket.Close(); err != nil {
			log.Println("failed to close socket:", err)
		}

		// Report any incomplete results (status reply but no name replies)
		p.ReportIncompleteResults()

		// Complete
		p.waiter.Done()
	}()

	return
}

func init() {
	probes = append(probes, new(ProbeNetbios))
}
