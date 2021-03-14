package commandline

import "time"

// CommandLine type represents options read from command line arguments.
type CommandLine struct {
	host    string
	port    uint64
	drop    string
	xtrn    string
	from    string
	timeout time.Duration
}

// Host method returns a given host.
func (c *CommandLine) Host() string {
	return c.host
}

// Port method returns a given port.
func (c *CommandLine) Port() uint64 {
	return c.port
}

// Drop method returns a given node.
func (c *CommandLine) Drop() string {
	return c.drop
}

// Xtrn method returns a given xtrn id.
func (c *CommandLine) Xtrn() string {
	return c.xtrn
}

// From method returns a given source BBS name.
func (c *CommandLine) From() string {
	return c.from
}

// Timeout method returns a given server response timeout after EOF of input file.
func (c *CommandLine) Timeout() time.Duration {
	return c.timeout
}
