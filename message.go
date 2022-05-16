package gomsgprocessor

import "context"

// Message is the input of ParallelProcessor's MakeDocuments.
type Message interface {
	GetNamespace() Namespace
	GetType() MessageType
	UpdateLogWithData(context.Context)
}
