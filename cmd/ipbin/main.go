package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"github.com/anatoly-kussul/ipbin/ipbin"
	"go4.org/netipx"
	"io"
	"net/netip"
	"os"
)

const (
	OutFormatSubnetsIPs = 1 + iota
	OutFormatRangesIPs
	OutFormatSubnets
	OutFormatRanges
)

type options struct {
	inputFilepath  string
	outputFilepath string
	gzipOut        bool
	gzipIn         bool
	binIn          bool
	binOut         bool
	sepOut         string // only if not binOut, separator for text output, \n by default
	formatOut      int    // only if not binOut
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: ipbin [options] <output-file>

Options:
  -i, --input string       Input file path
  -B                       Read input as binary
  -Z                       Read input as gzip
  -b                       Write output as binary
  -z                       Write output as gzip
  -s, --sep string         Separator for text output (default: \n)
  -f, --format int         Output format (1=subnets+ips, 2=ranges+ips, 3=subnets, 4=ranges)
  -h, --help               Show this help message
`)
}

// readPrefixes reads prefixes from the input file according to options
func readPrefixes(opts *options) ([]netip.Prefix, error) {
	var r io.Reader
	f, err := os.Open(opts.inputFilepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r = f
	if opts.gzipIn {
		gzr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		r = gzr
		defer gzr.Close()
	} else {
		r = bufio.NewReaderSize(r, 1024*32)
	}

	if opts.binIn {
		// Read all bytes, decode prefixes
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		var prefixes []netip.Prefix
		for len(data) > 0 {
			prefix, n, err := ipbin.ReadPrefixFromBytes(data)
			if err != nil {
				return nil, err
			}
			prefixes = append(prefixes, prefix)
			data = data[n:]
		}
		return prefixes, nil
	} else {
		return ipbin.ParseIPSubnets(r)
	}
}

// writePrefixes writes prefixes to the output file according to options
func writePrefixes(opts *options, ipset *netipx.IPSet) error {
	var w io.Writer
	f, err := os.Create(opts.outputFilepath)
	if err != nil {
		return err
	}
	defer f.Close()
	w = f
	if opts.gzipOut {
		gz := gzip.NewWriter(w)
		defer gz.Close()
		w = gz
	} else {
		bufw := bufio.NewWriterSize(w, 1024*32)
		defer bufw.Flush()
		w = bufw
	}

	if opts.binOut {
		for _, p := range ipset.Prefixes() {
			if _, err = ipbin.WriteEncoded(w, p); err != nil {
				return err
			}
		}
		return nil
	}

	// Text output with format
	sep := opts.sepOut

	switch opts.formatOut {
	case OutFormatSubnets:
		// Output merged subnets
		out := ipset.Prefixes()
		for i, p := range out {
			if i > 0 {
				if _, err = w.Write([]byte(sep)); err != nil {
					return err
				}
			}
			if _, err = w.Write([]byte(p.String())); err != nil {
				return err
			}
		}
	case OutFormatSubnetsIPs:
		// Output IP if prefix is a single IP, otherwise output prefix
		out := ipset.Prefixes()
		for i, p := range out {
			if i > 0 {
				if _, err = w.Write([]byte(sep)); err != nil {
					return err
				}
			}
			if p.Addr().BitLen() == p.Bits() {
				if _, err = w.Write([]byte(p.Addr().String())); err != nil {
					return err
				}
			} else {
				if _, err = w.Write([]byte(p.String())); err != nil {
					return err
				}
			}
		}
	case OutFormatRanges:
		// Output each range as start-end using ipset.Ranges()
		ranges := ipset.Ranges()
		for i, r := range ranges {
			if i > 0 {
				if _, err = w.Write([]byte(sep)); err != nil {
					return err
				}
			}
			if _, err = w.Write([]byte(r.From().String() + "-" + r.To().String())); err != nil {
				return err
			}
		}
	case OutFormatRangesIPs:
		// Output IP if range is a single IP, otherwise output range as start-end using ipset.Ranges()
		ranges := ipset.Ranges()
		for i, r := range ranges {
			if i > 0 {
				if _, err = w.Write([]byte(sep)); err != nil {
					return err
				}
			}
			if r.From() == r.To() {
				if _, err = w.Write([]byte(r.From().String())); err != nil {
					return err
				}
			} else {
				if _, err = w.Write([]byte(r.From().String() + "-" + r.To().String())); err != nil {
					return err
				}
			}
		}
	default:
		return fmt.Errorf("unknown output format: %d", opts.formatOut)
	}
	return nil
}

// expandShortFlags expands combined single-letter flags (e.g., -bz to -b -z)
func expandShortFlags(args []string) []string {
	var out []string
	for _, arg := range args {
		if len(arg) > 2 && arg[0] == '-' && arg[1] != '-' {
			// e.g., -bz -> -b -z
			for _, c := range arg[1:] {
				out = append(out, "-"+string(c))
			}
		} else {
			out = append(out, arg)
		}
	}
	return out
}

func main() {
	var opts options
	var showHelp bool

	flag.StringVar(&opts.inputFilepath, "input", "", "Input file path")
	flag.StringVar(&opts.inputFilepath, "i", "", "Input file path (shorthand)")
	flag.BoolVar(&opts.gzipIn, "Z", false, "Read input as gzip")
	flag.BoolVar(&opts.gzipOut, "z", false, "Write output as gzip")
	flag.BoolVar(&opts.binIn, "B", false, "Read input as binary")
	flag.BoolVar(&opts.binOut, "b", false, "Write output as binary")
	flag.StringVar(&opts.sepOut, "sep", "\n", "Separator for text output")
	flag.IntVar(&opts.formatOut, "format", OutFormatSubnetsIPs, "Output format (1=subnets, 2=subnets+ips, 3=ranges, 4=ranges+ips)")
	flag.IntVar(&opts.formatOut, "f", OutFormatSubnetsIPs, "Output format (shorthand)")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.BoolVar(&showHelp, "h", false, "Show help message (shorthand)")

	flag.Usage = usage
	// Expand combined short flags before parsing
	os.Args = expandShortFlags(os.Args)
	flag.Parse()

	if showHelp {
		usage()
		os.Exit(0)
	}

	// Output file is now a required positional argument
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: output file must be specified as a positional argument.\n")
		usage()
		os.Exit(2)
	}
	opts.outputFilepath = args[0]

	if opts.inputFilepath == "" || opts.outputFilepath == "" {
		fmt.Fprintf(os.Stderr, "Error: input and output file paths must be specified.\n")
		usage()
		os.Exit(2)
	}

	fmt.Printf("Reading input from %s...\n", opts.inputFilepath)
	prefixes, err := readPrefixes(&opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Merging prefixes...")
	ipset, err := ipbin.MergePrefixes(prefixes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error merging prefixes: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Writing output to %s...\n", opts.outputFilepath)
	if err := writePrefixes(&opts, ipset); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done.")
}
