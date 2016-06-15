package speed

// Writer defines the interface of a MMV file writer's properties
type Writer interface {
	Registry() *Registry // a writer must contain a registry of metrics and instance domains
	Write() error        // writes an mmv file
}
