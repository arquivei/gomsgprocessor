package gomsgprocessor

// Document is the result of a DocumentBuilder's Build
type Document interface{}

// Namespace is a logical separator that will be used to group messages while
// processing then
type Namespace string

// MessageType is used to decide which DocumentBuilder to use for each Message
type MessageType string
