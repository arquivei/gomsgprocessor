package gomsgprocessor

// DeduplicateDocumentsFunc is used to deduplicate a slice of Document.
type DeduplicateDocumentsFunc func([]Document) ([]Document, error)

func defaultDeduplicateDocumentsFunc(d []Document) ([]Document, error) {
	return d, nil
}
