package ports

// PortInfo holds enriched metadata about a single open port.
type PortInfo struct {
	Port    int
	Service string
	Process string
	PID     int
}

// Enricher combines resolver and process information to produce PortInfo.
type Enricher struct {
	resolver *Resolver
	procs    *ProcessResolver
}

// NewEnricher creates an Enricher backed by the given Resolver and ProcessResolver.
func NewEnricher(r *Resolver, p *ProcessResolver) *Enricher {
	return &Enricher{resolver: r, procs: p}
}

// Enrich returns a PortInfo slice for the provided port numbers.
func (e *Enricher) Enrich(ports []int) []PortInfo {
	procMap := map[int]ProcessInfo{}
	if e.procs != nil {
		if infos, err := e.procs.Resolve(); err == nil {
			for _, pi := range infos {
				procMap[pi.Port] = pi
			}
		}
	}

	result := make([]PortInfo, 0, len(ports))
	for _, p := range ports {
		info := PortInfo{Port: p}
		if e.resolver != nil {
			info.Service = e.resolver.Lookup(p)
		}
		if pi, ok := procMap[p]; ok {
			info.Process = pi.Name
			info.PID = pi.PID
		}
		result = append(result, info)
	}
	return result
}
