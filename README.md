# ipbin

is a command-line utility designed for efficiently storing large lists of IP addresses and subnets using a compact custom binary format, with additional support for merging, converting, and transforming IP data from various formats.
## Installation

via go install:
```bash
$ go install github.com/anatoly-kussul/ipbin/cmd/ipbin@latest
```
This installs ipbin to $GOPATH/bin/ipbin

## Usage

```
ipbin [options] <output-file>
```

### Options

```
  -i, --input string      Input file path
  -B                      Read input as binary
  -Z                      Read input as gzip
  -b                      Write output as binary
  -z                      Write output as gzip
  -s, --sep string        Separator for text output (default: \n)
  -f, --format int        Text output format (1=subnets+ips, 2=ranges+ips, 3=subnets, 4=ranges)
  -h, --help              Show this help message
```

### Binary Output Format
If `-b` is specified, output is written in a compact binary format:
- Each prefix is encoded as follows:
  - The first byte encodes both the address family and prefix length:
    - Values 0–32 represent an IPv4 prefix length.
    - Values 33–161 represent an IPv6 prefix length (actual length = byte - 33).
  - The following bytes contain only the minimum number of bytes required to represent the prefix address, i.e., ceil(prefixLen / 8) bytes.
- Example:
  - IPv4 /24 → b[0] = 24, b[1:4] = first 3 bytes of IPv4 address
  - IPv6 /64 → b[0] = 97 (64 + 33), b[1:9] = first 8 bytes of IPv6 address
- The file is a concatenation of such encoded prefixes.

### Text Output Formats
- `1` (default): subnets+ips — single IPs as IPs, others as subnets
- `2`: ranges+ips — single IPs as IPs, others as start-end
- `3`: subnets — subnets only
- `4`: ranges — ranges as start-end only

## Input Format
- Text input: one IP, subnet, or range per line (e.g., `1.2.3.4`, `10.0.0.0/8`, `192.168.1.1-192.168.1.255`)
- Binary input: compact encoded prefixes as described above

## License
MIT