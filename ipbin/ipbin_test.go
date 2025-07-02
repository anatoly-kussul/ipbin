package ipbin

import (
	"bytes"
	"net/netip"
	"testing"
)

type testCase struct {
	p netip.Prefix
	b []byte
}

var cases = []testCase{
	{netip.MustParsePrefix("1.2.3.0/31"), []byte{31, 1, 2, 3, 0}},
	{netip.MustParsePrefix("1.3.0.0/16"), []byte{16, 1, 3}},
	{netip.MustParsePrefix("1.4.0.0/23"), []byte{23, 1, 4, 0}},
	{netip.MustParsePrefix("1.5.5.5/32"), []byte{32, 1, 5, 5, 5}},
	{
		netip.MustParsePrefix("2003:c1:c72e:6100:f921:95a6:ef39:1ec7/128"),
		[]byte{128 + 33, 0x20, 0x03, 0x00, 0xc1, 0xc7, 0x2e, 0x61, 0x00, 0xf9, 0x21, 0x95, 0xa6, 0xef, 0x39, 0x1e, 0xc7},
	},
	{
		netip.MustParsePrefix("2003:c1:c72e:6100:f921:95a6:ef39:2ec8/127"),
		[]byte{127 + 33, 0x20, 0x03, 0x00, 0xc1, 0xc7, 0x2e, 0x61, 0x00, 0xf9, 0x21, 0x95, 0xa6, 0xef, 0x39, 0x2e, 0xc8},
	},
	{netip.MustParsePrefix("2001:db8:abcd:1234::/64"), []byte{64 + 33, 0x20, 0x01, 0x0d, 0xb8, 0xab, 0xcd, 0x12, 0x34}},
}

func TestEncodePrefix(t *testing.T) {
	for _, tc := range cases {
		b, n, err := EncodePrefix(tc.p)
		if err != nil {
			t.Errorf("EncodePrefix(%#v) error %v", tc.p, err)
			return
		}
		if !bytes.Equal(b[:n], tc.b) {
			t.Errorf("EncodePrefix(%#v) got %#v, want %#v", tc.p, b[:n], tc.b)
			return
		}
	}
}

func TestDecodePrefix(t *testing.T) {
	var buf []byte
	for _, tc := range cases {
		buf = append(buf, tc.b...)
	}
	for i := 0; len(buf) > 0; i++ {
		prefix, n, err := ReadPrefixFromBytes(buf)
		if err != nil {
			t.Errorf("ReadPrefixFromBytes error %v", err)
			return
		}
		buf = buf[n:]
		if prefix != cases[i].p {
			t.Errorf("ReadPrefixFromBytes got %#v, want %#v", prefix, cases[i].p)
			return
		}
	}
}
