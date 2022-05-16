package gomsgprocessor

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/arquivei/foundationkit/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_MakeDocuments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string

		messages                 []Message
		builders                 map[MessageType]DocumentBuilder
		deduplicateDocumentsFunc DeduplicateDocumentsFunc

		expectedResponse  []Document
		expectedError     string
		expectedErrorCode errors.Code
	}{
		{
			name: "success - 1 message - 1 document",
			messages: []Message{
				&mockMessage{
					id:          "id-1",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
			},
			builders: map[MessageType]DocumentBuilder{
				MessageType("type-1"): func() DocumentBuilder {
					m := new(mockDocumentBuilder)
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-1",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-1"},
							},
							nil,
						).
						Once()
					return m
				}(),
			},
			expectedResponse: []Document{
				mockDocument{id: "doc-1"},
			},
		},
		{
			name: "success - 1 message - 1 noop",
			messages: []Message{
				&mockMessage{
					id:          "id-1",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
			},
			builders: map[MessageType]DocumentBuilder{
				MessageType("type-1"): func() DocumentBuilder {
					m := new(mockDocumentBuilder)
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-1",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							nil,
							nil,
						).
						Once()
					return m
				}(),
			},
			expectedResponse: []Document{},
		},
		{
			name: "success - 1 message - 2 documents",
			messages: []Message{
				&mockMessage{
					id:          "id-1",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
			},
			builders: map[MessageType]DocumentBuilder{
				MessageType("type-1"): func() DocumentBuilder {
					m := new(mockDocumentBuilder)
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-1",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-1"},
								mockDocument{id: "doc-2"},
							},
							nil,
						).
						Once()
					return m
				}(),
			},
			expectedResponse: []Document{
				mockDocument{id: "doc-1"},
				mockDocument{id: "doc-2"},
			},
		},
		{
			name: "success - 3 messages - 6 documents",
			messages: []Message{
				&mockMessage{
					id:          "id-1",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-2",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-3",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
			},
			builders: map[MessageType]DocumentBuilder{
				MessageType("type-1"): func() DocumentBuilder {
					m := new(mockDocumentBuilder)
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-1",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-1"},
								mockDocument{id: "doc-2"},
							},
							nil,
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-2",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-3"},
								mockDocument{id: "doc-4"},
							},
							nil,
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-3",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-5"},
								mockDocument{id: "doc-6"},
							},
							nil,
						).
						Once()
					return m
				}(),
			},
			expectedResponse: []Document{
				mockDocument{id: "doc-1"},
				mockDocument{id: "doc-2"},
				mockDocument{id: "doc-3"},
				mockDocument{id: "doc-4"},
				mockDocument{id: "doc-5"},
				mockDocument{id: "doc-6"},
			},
		},
		{
			name: "success - 3 messages - 4 documents - 1 noop",
			messages: []Message{
				&mockMessage{
					id:          "id-1",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-2",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-3",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
			},
			builders: map[MessageType]DocumentBuilder{
				MessageType("type-1"): func() DocumentBuilder {
					m := new(mockDocumentBuilder)
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-1",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-1"},
								mockDocument{id: "doc-2"},
							},
							nil,
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-2",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-3"},
								mockDocument{id: "doc-4"},
							},
							nil,
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-3",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							nil,
							nil,
						).
						Once()
					return m
				}(),
			},
			expectedResponse: []Document{
				mockDocument{id: "doc-1"},
				mockDocument{id: "doc-2"},
				mockDocument{id: "doc-3"},
				mockDocument{id: "doc-4"},
			},
		},
		{
			name: "success - 3 messages - 2 builders - 4 documents - 1 noop",
			messages: []Message{
				&mockMessage{
					id:          "id-1",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-2",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-3",
					namespace:   "tiramisu",
					messageType: "type-2",
				},
			},
			builders: map[MessageType]DocumentBuilder{
				MessageType("type-1"): func() DocumentBuilder {
					m := new(mockDocumentBuilder)
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-1",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-1"},
								mockDocument{id: "doc-2"},
							},
							nil,
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-2",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-3"},
								mockDocument{id: "doc-4"},
							},
							nil,
						).
						Once()
					return m
				}(),
				MessageType("type-2"): func() DocumentBuilder {
					m := new(mockDocumentBuilder)
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-3",
								namespace:   "tiramisu",
								messageType: "type-2",
							},
						).
						Return(
							nil,
							nil,
						).
						Once()
					return m
				}(),
			},
			expectedResponse: []Document{
				mockDocument{id: "doc-1"},
				mockDocument{id: "doc-2"},
				mockDocument{id: "doc-3"},
				mockDocument{id: "doc-4"},
			},
		},
		{
			name: "success - 4 messages - 2 namespaces",
			messages: []Message{
				&mockMessage{
					id:          "id-1",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-2",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-1",
					namespace:   "potato",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-2",
					namespace:   "potato",
					messageType: "type-1",
				},
			},
			builders: map[MessageType]DocumentBuilder{
				MessageType("type-1"): func() DocumentBuilder {
					m := new(mockDocumentBuilder)
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-1",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-1"},
								mockDocument{id: "doc-2"},
							},
							nil,
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-2",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-3"},
								mockDocument{id: "doc-4"},
							},
							nil,
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-1",
								namespace:   "potato",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-1"},
								mockDocument{id: "doc-2"},
							},
							nil,
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-2",
								namespace:   "potato",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-3"},
								mockDocument{id: "doc-4"},
							},
							nil,
						).
						Once()
					return m
				}(),
			},
			expectedResponse: []Document{
				mockDocument{id: "doc-1"},
				mockDocument{id: "doc-1"},
				mockDocument{id: "doc-2"},
				mockDocument{id: "doc-2"},
				mockDocument{id: "doc-3"},
				mockDocument{id: "doc-3"},
				mockDocument{id: "doc-4"},
				mockDocument{id: "doc-4"},
			},
		},
		{
			name: "errors - document builder",
			messages: []Message{
				&mockMessage{
					id:          "id-1",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-2",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-3",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
			},
			builders: map[MessageType]DocumentBuilder{
				MessageType("type-1"): func() DocumentBuilder {
					m := new(mockDocumentBuilder)
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-1",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-1"},
								mockDocument{id: "doc-2"},
							},
							nil,
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-2",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-3"},
								mockDocument{id: "doc-4"},
							},
							errors.New("document builder error"),
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-3",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							nil,
							nil,
						).
						Once()
					return m
				}(),
			},
			expectedError:     "gomsgprocessor.parallelProcessor.MakeDocuments: parallelBuildDocumentsByNamespace: document builder error",
			expectedErrorCode: ErrCodeBuildDocuments,
		},
		{
			name: "errors - deduplicate",
			messages: []Message{
				&mockMessage{
					id:          "id-1",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-2",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
				&mockMessage{
					id:          "id-3",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
			},
			builders: map[MessageType]DocumentBuilder{
				MessageType("type-1"): func() DocumentBuilder {
					m := new(mockDocumentBuilder)
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-1",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-1"},
								mockDocument{id: "doc-2"},
							},
							nil,
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-2",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							[]Document{
								mockDocument{id: "doc-3"},
								mockDocument{id: "doc-4"},
							},
							nil,
						).
						Once()
					m.
						On(
							"Build",
							&mockMessage{
								id:          "id-3",
								namespace:   "tiramisu",
								messageType: "type-1",
							},
						).
						Return(
							nil,
							nil,
						).
						Once()
					return m
				}(),
			},
			deduplicateDocumentsFunc: mockDeduplicateDocumentsFuncError,
			expectedError:            "gomsgprocessor.parallelProcessor.MakeDocuments: deduplicateDocumentsForEachNamespace: deduplicate documents mock error",
			expectedErrorCode:        ErrCodeDeduplicateDocuments,
		},
		{
			name: "errors - no document builder for message",
			messages: []Message{
				&mockMessage{
					id:          "id-1",
					namespace:   "tiramisu",
					messageType: "type-1",
				},
			},
			builders:          map[MessageType]DocumentBuilder{},
			expectedError:     "gomsgprocessor.parallelProcessor.MakeDocuments: parallelBuildDocumentsByNamespace: message type has no document builder [type=type-1]",
			expectedErrorCode: ErrCodeBuildDocuments,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parallelProcessor := NewParallelProcessor(
				test.builders,
			)

			if test.deduplicateDocumentsFunc != nil {
				parallelProcessor = NewParallelProcessor(
					test.builders,
					WithDeduplicateDocumentsOption(test.deduplicateDocumentsFunc),
				)
			}

			documents, err := parallelProcessor.MakeDocuments(
				context.Background(),
				test.messages,
			)

			if documents != nil {
				// for test purpose
				sortByID(documents)
			}

			if test.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedResponse, documents)
			} else {
				assert.EqualError(t, err, test.expectedError)
				assert.Equal(t, test.expectedErrorCode, errors.GetCode(err))
			}

			for _, builder := range test.builders {
				builder.(*mockDocumentBuilder).AssertExpectations(t)
			}
		})
	}
}

func sortByID(docs []Document) []Document {
	mockDocuments := make([]mockDocument, 0, len(docs))
	for _, doc := range docs {
		mockDocuments = append(mockDocuments, doc.(mockDocument))
	}
	sort.SliceStable(mockDocuments, func(i, j int) bool {
		return strings.Compare(mockDocuments[i].id, mockDocuments[j].id) < 0
	})
	for i := range docs {
		docs[i] = mockDocuments[i]
	}
	return docs
}

type mockDocumentBuilder struct {
	mock.Mock
}

func (g *mockDocumentBuilder) Build(
	_ context.Context,
	m Message,
) ([]Document, error) {
	args := g.Called(m)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Document), args.Error(1)
}

type mockMessage struct {
	id          string
	namespace   Namespace
	messageType MessageType
}

func (m *mockMessage) GetNamespace() Namespace {
	return m.namespace
}

func (m *mockMessage) GetType() MessageType {
	return m.messageType
}

func (m *mockMessage) UpdateLogWithData(context.Context) {
}

type mockDocument struct {
	id string
}

func mockDeduplicateDocumentsFuncError(d []Document) ([]Document, error) {
	return nil, errors.New("deduplicate documents mock error")
}
