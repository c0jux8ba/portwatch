package ports

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ServiceName maps a port number to a well-known service name.
// Falls back to "/etc/services" on Unix systems, then to a small
// built-in table so the feature works on all platforms.
type Resolver struct {
	table map[int]string
}

// builtinServices is a minimal fallback table.
var builtinServices = map[int]string{
	21:   "ftp",
	22:   "ssh",
	23:   "telnet",
	25:   "smtp",
	53:   "dns",
	80:   "http",
	110:  "pop3",
	143:  "imap",
	443:  "https",
	3306: "mysql",
	5432: "postgres",
	6379: "redis",
	8080: "http-alt",
	27017: "mongodb",
}

// NewResolver builds a Resolver, attempting to parse /etc/services first.
func NewResolver() *Resolver {
	r := &Resolver{table: make(map[int]string)}
	for k, v := range builtinServices {
		r.table[k] = v
	}
	r.loadEtcServices("/etc/services")
	return r
}

// Lookup returns the service name for port, or a numeric string if unknown.
func (r *Resolver) Lookup(port int) string {
	if name, ok := r.table[port]; ok {
		return name
	}
	return strconv.Itoa(port)
}

// LookupAll resolves a slice of port numbers to "port/name" strings.
func (r *Resolver) LookupAll(ports []int) []string {
	out := make([]string, len(ports))
	for i, p := range ports {
		out[i] = fmt.Sprintf("%d/%s", p, r.Lookup(p))
	}
	return out
}

func (r *Resolver) loadEtcServices(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		portProto := strings.SplitN(fields[1], "/", 2)
		if len(portProto) < 2 || portProto[1] != "tcp" {
			continue
		}
		port, err := strconv.Atoi(portProto[0])
		if err != nil {
			continue
		}
		if _, exists := r.table[port]; !exists {
			r.table[port] = name
		}
	}
}
