package ipbin

import (
	"net/netip"
	"reflect"
	"strings"
	"testing"
)

func TestParseIPSubnets(t *testing.T) {
	input := `1.2.3.0
	1.2.3.1
	1.4.0.0-1.4.0.255
	1.4.1.0-1.4.1.255
	1.3.1.0-1.3.1.255
	1.3.0.0-1.3.255.255
	1.3.100.200
	1.5.5.5
	2003:c1:c72e:6100:f921:95a6:ef39:1ec7
	2003:c1:c72e:6100:f921:95a6:ef39:2ec8-2003:c1:c72e:6100:f921:95a6:ef39:2ec9
	10.0.0.0/16`
	r := strings.NewReader(input)
	nets, err := ParseIPSubnets(r)
	if err != nil {
		t.Error(err)
		return
	}
	ipset, err := MergePrefixes(nets)
	if err != nil {
		t.Error(err)
		return
	}
	nets = ipset.Prefixes()
	expected := []netip.Prefix{
		netip.MustParsePrefix("1.2.3.0/31"),
		netip.MustParsePrefix("1.3.0.0/16"),
		netip.MustParsePrefix("1.4.0.0/23"),
		netip.MustParsePrefix("1.5.5.5/32"),
		netip.MustParsePrefix("10.0.0.0/16"),
		netip.MustParsePrefix("2003:c1:c72e:6100:f921:95a6:ef39:1ec7/128"),
		netip.MustParsePrefix("2003:c1:c72e:6100:f921:95a6:ef39:2ec8/127"),
	}
	if !reflect.DeepEqual(nets, expected) {
		t.Errorf("got %v\nwant %v", nets, expected)
		return
	}
}
