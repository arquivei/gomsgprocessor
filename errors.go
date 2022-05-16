package gomsgprocessor

import "github.com/arquivei/foundationkit/errors"

// ErrCodeBuildDocuments is returned when the operation failed to build the
// response for given []Message
var ErrCodeBuildDocuments = errors.Code("FAILED_BUILD_DOCUMENTS")

// ErrCodeDeduplicateDocuments is returned when the operation failed to
// deduplicate []Documents by Namespace
var ErrCodeDeduplicateDocuments = errors.Code("FAILED_DEDUPLICATE_DOCUMENTS")

// ErrMsgTypeHasNoBuilder is returned when MessageType has no DocumentBuilder
// associated with it
var ErrMsgTypeHasNoBuilder = errors.New("message type has no document builder")
