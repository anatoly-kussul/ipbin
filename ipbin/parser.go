package ipbin

import (
	"bufio"
	"go4.org/netipx"
	"io"
	"net/netip"
	"strings"
)

func ParseIPSubnets(r io.Reader) (nets []netip.Prefix, err error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		s := strings.Split(line, ",")[0]
		switch {
		case strings.Contains(s, "-"):
			rangeS := strings.Split(s, "-")
			startIp, err := netip.ParseAddr(strings.TrimSpace(rangeS[0]))
			if err != nil {
				return nil, err
			}
			if len(s) > 1 {
				endIp, err := netip.ParseAddr(strings.TrimSpace(rangeS[1]))
				if err != nil {
					return nil, err
				}
				nets = netipx.IPRangeFrom(startIp, endIp).AppendPrefixes(nets)
			} else {
				nets = append(nets, netip.PrefixFrom(startIp, startIp.BitLen()))
			}
		case strings.Contains(s, "/"):
			prefix, err := netip.ParsePrefix(strings.TrimSpace(s))
			if err != nil {
				return nil, err
			}
			nets = append(nets, prefix)
		default:
			ip, err := netip.ParseAddr(strings.TrimSpace(s))
			if err != nil {
				return nil, err
			}
			nets = append(nets, netip.PrefixFrom(ip, ip.BitLen()))
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return nets, nil
}

// MergePrefixes takes a slice of netip.Prefix values and returns a new slice
// where all adjacent or overlapping prefixes have been merged into the minimal
// set of covering prefixes.
//
// For example, if two prefixes 192.168.1.0/25 and 192.168.1.128/25 are given,
// they will be merged into a single prefix 192.168.1.0/24.
//
// The function does not modify the input slice. The result is sorted and
// non-overlapping.
func MergePrefixes(prefixes []netip.Prefix) (*netipx.IPSet, error) {
	builder := netipx.IPSetBuilder{}
	for _, prefix := range prefixes {
		builder.AddPrefix(prefix)
	}
	return builder.IPSet()
}
