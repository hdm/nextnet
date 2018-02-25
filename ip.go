package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"regexp"
	"strings"
)

// MatchIPv6 is a regular expression for matching IPv6 addresses.
var MatchIPv6 = regexp.MustCompile(`^((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?$`)

// MatchIPv4 is a regular expression for matching IPv4 addresses.
var MatchIPv4 = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2}))$`)

// IPv4Masks is a map with key as mask size and value as number of mask bits.
var IPv4Masks = map[uint32]uint32{
	1:          32,
	2:          31,
	4:          30,
	8:          29,
	16:         28,
	32:         27,
	64:         26,
	128:        25,
	256:        24,
	512:        23,
	1024:       22,
	2048:       21,
	4096:       20,
	8192:       19,
	16384:      18,
	32768:      17,
	65536:      16,
	131072:     15,
	262144:     14,
	524288:     13,
	1048576:    12,
	2097152:    11,
	4194304:    10,
	8388608:    9,
	16777216:   8,
	33554432:   7,
	67108864:   6,
	134217728:  5,
	268435456:  4,
	536870912:  3,
	1073741824: 2,
	2147483648: 1,
}

// IPv4MaskSizes is a slice of all IPv4 mask sizes.
var IPv4MaskSizes = []uint32{
	2147483648,
	1073741824,
	536870912,
	268435456,
	134217728,
	67108864,
	33554432,
	16777216,
	8388608,
	4194304,
	2097152,
	1048576,
	524288,
	262144,
	131072,
	65536,
	32768,
	16384,
	8192,
	4096,
	2048,
	1024,
	512,
	256,
	128,
	64,
	32,
	16,
	8,
	4,
	2,
	1,
}

// IPv4ToUint converts string IP to uint32.
func IPv4ToUint(ips string) (uint32, error) {
	ip := net.ParseIP(ips)
	if ip == nil {
		return 0, errors.New("Invalid IPv4 address")
	}

	ip = ip.To4()
	return binary.BigEndian.Uint32(ip), nil
}

// UintToIPv4 converts uint32 IP to string.
func UintToIPv4(ipi uint32) string {
	ipb := make([]byte, 4)
	binary.BigEndian.PutUint32(ipb, ipi)
	ip := net.IP(ipb)
	return ip.String()
}

// IPv4Range2CIDRs returns range of CIDRs built by string IPs.
func IPv4Range2CIDRs(startIP, endIP string) ([]string, error) {
	start, err := IPv4ToUint(startIP)
	if err != nil {
		return []string{}, err
	}

	end, err := IPv4ToUint(endIP)
	if err != nil {
		return []string{}, err
	}

	if start > end {
		return []string{}, errors.New("Start address is bigger than end address")
	}

	return IPv4UintRange2CIDRs(start, end), nil
}

// IPv4UintRange2CIDRs returns range of CIDRs built by uint32 IPs.
func IPv4UintRange2CIDRs(startIP, endIP uint32) (cidrs []string) {
	// Ranges are inclusive
	size := endIP - startIP + 1

	if size == 0 {
		return cidrs
	}

	for i := range IPv4MaskSizes {

		maskSize := IPv4MaskSizes[i]

		if maskSize > size {
			continue
		}

		// Exact match of the block size
		if maskSize == size {
			cidrs = append(cidrs, fmt.Sprintf("%s/%d", UintToIPv4(startIP), IPv4Masks[maskSize]))
			break
		}

		// Chop off the biggest block that fits
		cidrs = append(cidrs, fmt.Sprintf("%s/%d", UintToIPv4(endIP), IPv4Masks[maskSize]))
		startIP += maskSize

		// Look for additional blocks
		newCIDRs := IPv4UintRange2CIDRs(startIP, endIP)

		// Merge those blocks into out results
		for x := range newCIDRs {
			cidrs = append(cidrs, newCIDRs[x])
		}
		break

	}

	return
}

// AddressesFromCIDR extracts addresses from CIDR and send its to out channel.
func AddressesFromCIDR(cidr string, out chan<- string) {
	if len(cidr) == 0 {
		return
	}

	// We may receive bare IP addresses, not CIDRs sometimes
	if !strings.Contains(cidr, "/") {
		if strings.Contains(cidr, ":") {
			cidr = cidr + "/128"
		} else {
			cidr = cidr + "/32"
		}
	}

	// Parse CIDR into base address + mask
	ip, net, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid CIDR %s: %s\n", cidr, err.Error())
		return
	}

	// Verify IPv4 for now
	ip4 := net.IP.To4()
	if ip4 == nil {
		fmt.Fprintf(os.Stderr, "Invalid IPv4 CIDR %s\n", cidr)
		return
	}

	netBase, err := IPv4ToUint(net.IP.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid IPv4 Address %s: %s\n", ip.String(), err.Error())
		return
	}

	maskOnes, maskTotal := net.Mask.Size()

	// Does not work for IPv6 due to cast to uint32
	netSize := uint32(math.Pow(2, float64(maskTotal-maskOnes)))

	curBase := netBase
	endBase := netBase + netSize
	curAddr := curBase

	for curAddr = curBase; curAddr < endBase; curAddr++ {
		out <- UintToIPv4(curAddr)
	}

	return
}
