package ipbin

import (
	"fmt"
	"io"
	"net/netip"
)

// EncodePrefix encodes a netip.Prefix into a compact binary representation.
//
// The encoding format is:
//   - The first byte encodes both the address family and prefix length:
//     Values 0–32 represent an IPv4 prefix length.
//     Values 33–161 represent an IPv6 prefix length (actual length = byte - 33).
//   - The following bytes contain only the minimum number of bytes required
//     to represent the prefix address, i.e., ceil(prefixLen / 8) bytes.
//
// The function returns:
//   - b: a fixed-size 17-byte array containing the encoded prefix.
//   - n: the actual number of bytes used in b.
//   - err: an error if the prefix is invalid.
//
// Example:
//   - IPv4 /24 → b[0] = 24, b[1:4] = first 3 bytes of IPv4 address, n = 4
//   - IPv6 /64 → b[0] = 97 (64 + 33), b[1:9] = first 8 bytes of IPv6 address, n = 9
func EncodePrefix(p netip.Prefix) (b [17]byte, n int, err error) {
	if !p.IsValid() {
		err = fmt.Errorf("invalid preifx %v", p)
		return
	}
	addr := p.Addr()
	bits := p.Bits()
	prefixBytesLen := (bits + 7) / 8
	if addr.Is4() {
		b[0] = byte(bits)
		ip := addr.As4()
		copy(b[1:], ip[:prefixBytesLen])
	} else {
		b[0] = byte(bits + 33)
		ip := addr.As16()
		copy(b[1:], ip[:prefixBytesLen])
	}
	n = prefixBytesLen + 1
	return
}

func WriteEncoded(w io.Writer, p netip.Prefix) (n int, err error) {
	b, n, err := EncodePrefix(p)
	if err != nil {
		return 0, err
	}
	return w.Write(b[:n])
}

func AppendEncoded(dst []byte, p netip.Prefix) ([]byte, error) {
	b, n, err := EncodePrefix(p)
	if err != nil {
		return nil, err
	}
	return append(dst, b[:n]...), nil
}

// ReadPrefixFromBytes reads from buf and returns netip.Prefix, int of bytes read and/or error
//
// Example usage:
//
//	f, err := os.Open("prefixes.bin")
//	if err != nil {
//	    fmt.Println("Failed to open file:", err)
//	    return
//	}
//
//	data, err := io.ReadAll(f)
//	f.Close()
//	if err != nil {
//	    fmt.Println("Failed to read file:", err)
//	    return
//	}
//
//	for len(data) > 0 {
//	    prefix, n, err := ReadPrefixFromBytes(data)
//	    if err != nil {
//	        fmt.Println("Error decoding prefix:", err)
//	        break
//	    }
//	    fmt.Println("Prefix:", prefix)
//	    data = data[n:]
//	}
func ReadPrefixFromBytes(buf []byte) (netip.Prefix, int, error) {
	if len(buf) == 0 {
		return netip.Prefix{}, 0, io.EOF
	}

	hdr := buf[0]
	switch {
	case hdr <= 32: // IPv4
		prefixLen := int(hdr)
		numBytes := 1 + (prefixLen+7)/8
		if len(buf) < numBytes {
			return netip.Prefix{}, 0, io.ErrUnexpectedEOF
		}
		var ipv4 [4]byte
		copy(ipv4[:], buf[1:numBytes])
		prefix := netip.PrefixFrom(netip.AddrFrom4(ipv4), prefixLen)
		return prefix, numBytes, nil

	case hdr <= 161: // IPv6
		prefixLen := int(hdr) - 33
		numBytes := 1 + (prefixLen+7)/8
		if len(buf) < numBytes {
			return netip.Prefix{}, 0, io.ErrUnexpectedEOF
		}
		var ipv6 [16]byte
		copy(ipv6[:], buf[1:numBytes])
		prefix := netip.PrefixFrom(netip.AddrFrom16(ipv6), prefixLen)
		return prefix, numBytes, nil

	default:
		return netip.Prefix{}, 0, fmt.Errorf("invalid prefix header byte %d", hdr)
	}
}
