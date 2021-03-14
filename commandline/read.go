package commandline

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Read method returns valid options read from command line args.
func Read() *CommandLine {
	host := kingpin.Arg("host", "Target host").Required().String()
	port := kingpin.Arg("port", "Target port").Required().Uint64()
	drop := kingpin.Arg("node", "Target node").Required().String()
	xtrn := kingpin.Arg("xtrn", "Target xtrn").Required().String()
	from := kingpin.Arg("from", "Target from").Required().String()
	timeout := kingpin.Flag("timeout", "Byte receiving timeout after the input EOF occurs").Short('t').Default("1s").Duration()

	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("1.0").Author("Forked by Alpha from Marcin Tojek")
	kingpin.CommandLine.Name = "go-telnet"
	kingpin.CommandLine.Help = "Read bytes from stdin and pass them to the remote host."

	kingpin.Parse()

	return &CommandLine{
		host:    *host,
		port:    *port,
		drop:    *drop,
		xtrn:    *xtrn,
		from:    *from,
		timeout: *timeout,
	}
}

// SetCommandLineArgs method changes earlier set arguments (use only in debug).
func SetCommandLineArgs(customArguments ...string) {
	os.Args = os.Args[0:1] // leave only app path
	for _, customArgument := range customArguments {
		os.Args = append(os.Args, customArgument)
	}
}
