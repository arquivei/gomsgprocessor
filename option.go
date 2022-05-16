package gomsgprocessor

// Option is used to configure the processor
type Option func(*parallelProcessor)

// WithDeduplicateDocumentsOption adds a DeduplicateDocumentsFunc in
// ParallelProcessor, used to deduplicate a slice of Document
func WithDeduplicateDocumentsOption(d DeduplicateDocumentsFunc) Option {
	return func(p *parallelProcessor) {
		p.deduplicateDocuments = d
	}
}
