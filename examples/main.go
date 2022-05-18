// nolint
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gomsgprocessor"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ExampleMessage struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Age           int      `json:"age"`
	City          string   `json:"city"`
	State         string   `json:"state"`
	ChildrenNames []string `json:"childrenNames"`
	Namespace     string   `json:"namespace"`
}

func (e *ExampleMessage) GetNamespace() gomsgprocessor.Namespace {
	return gomsgprocessor.Namespace(e.Namespace)
}

func (e *ExampleMessage) GetType() gomsgprocessor.MessageType {
	return gomsgprocessor.MessageType("typeExample")
}

func (e *ExampleMessage) UpdateLogWithData(ctx context.Context) {
	log.Ctx(ctx).UpdateContext(func(zc zerolog.Context) zerolog.Context {
		return zc.
			Int("msg_id", e.ID).
			Str("msg_name", e.Name).
			Strs("msg_children_names", e.ChildrenNames).
			Int("msg_age", e.Age).
			Str("msg_city", e.City).
			Str("msg_state", e.State).
			Str("msg_type", string(e.GetType())).
			Str("msg_namespace", e.Namespace)
	})
}

type ExampleDocument struct {
	ID              string
	CreatedAt       time.Time
	ParentName      string
	ParentBirthYear int
	ChildName       string
	CityAndState    string
	Namespace       string
}

type ExampleBuilder struct{}

func (b *ExampleBuilder) Build(_ context.Context, msg gomsgprocessor.Message) ([]gomsgprocessor.Document, error) {
	exampleMsg, ok := msg.(*ExampleMessage)
	if !ok {
		return nil, errors.New("failed to cast message")
	}

	// Parallel Processor will ignore this message
	if len(exampleMsg.ChildrenNames) == 0 {
		return nil, nil
	}

	documents := make([]gomsgprocessor.Document, 0, len(exampleMsg.ChildrenNames))

	for _, childName := range exampleMsg.ChildrenNames {
		documents = append(documents, ExampleDocument{
			ID:              strconv.Itoa(exampleMsg.ID) + "_" + childName,
			CreatedAt:       time.Now(),
			ParentName:      exampleMsg.Name,
			CityAndState:    exampleMsg.City + " - " + exampleMsg.State,
			ChildName:       childName,
			ParentBirthYear: time.Now().Year() - exampleMsg.Age,
			Namespace:       exampleMsg.Namespace,
		})
	}

	return documents, nil
}

func ExampleDeduplicateDocuments(documents []gomsgprocessor.Document) ([]gomsgprocessor.Document, error) {
	examplesDocuments := make([]ExampleDocument, 0, len(documents))
	for _, document := range documents {
		exampleDocument, ok := document.(ExampleDocument)
		if !ok {
			return nil, errors.New("failed to cast document")
		}
		examplesDocuments = append(examplesDocuments, exampleDocument)
	}

	documentsByID := make(map[string]ExampleDocument, len(examplesDocuments))
	for _, exampleDocument := range examplesDocuments {
		documentsByID[exampleDocument.ID] = exampleDocument
	}

	deduplicatedDocuments := make([]gomsgprocessor.Document, 0, len(documentsByID))
	for _, documentByID := range documentsByID {
		deduplicatedDocuments = append(deduplicatedDocuments, documentByID)
	}
	return deduplicatedDocuments, nil
}

func main() {
	parallelProcessor := gomsgprocessor.NewParallelProcessor(
		map[gomsgprocessor.MessageType]gomsgprocessor.DocumentBuilder{
			"typeExample": &ExampleBuilder{},
		},
		gomsgprocessor.WithDeduplicateDocumentsOption(ExampleDeduplicateDocuments),
	)

	messages := []gomsgprocessor.Message{
		&ExampleMessage{
			ID:            1,
			Name:          "John",
			Age:           30,
			City:          "New York",
			State:         "NY",
			ChildrenNames: []string{"John", "Jane", "Mary"},
			Namespace:     "namespace1",
		},
		&ExampleMessage{
			ID:            2,
			Name:          "Poul",
			Age:           25,
			City:          "New Jersey",
			State:         "NY",
			ChildrenNames: []string{},
			Namespace:     "namespace1",
		},
		&ExampleMessage{
			ID:            3,
			Name:          "Chris",
			Age:           35,
			City:          "Washington",
			State:         "DC",
			ChildrenNames: []string{"Bob"},
			Namespace:     "namespace1",
		},
		&ExampleMessage{
			ID:            3,
			Name:          "Chris",
			Age:           35,
			City:          "Washington",
			State:         "DC",
			ChildrenNames: []string{"Bob"},
			Namespace:     "namespace2",
		},
		&ExampleMessage{
			ID:            1,
			Name:          "John",
			Age:           30,
			City:          "New York",
			State:         "NY",
			ChildrenNames: []string{"John", "Jane", "Mary"},
			Namespace:     "namespace1",
		},
	}

	documents, err := parallelProcessor.MakeDocuments(context.Background(), messages)
	if err != nil {
		panic(err)
	}

	examplesDocuments := make([]ExampleDocument, 0, len(documents))
	for _, document := range documents {
		exampleDocument, ok := document.(ExampleDocument)
		if !ok {
			panic("failed to cast document")
		}
		examplesDocuments = append(examplesDocuments, exampleDocument)
	}

	fmt.Println(JSONMarshal(examplesDocuments))
}

func JSONMarshal(t interface{}) (string, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(t)
	return buffer.String(), err
}

/*
Expected output:

[
  {
    "ID": "1_Jane",
    "CreatedAt": "2022-05-18T11:49:22.278074959-03:00",
    "ParentName": "John",
    "ParentBirthYear": 1992,
    "ChildName": "Jane",
    "CityAndState": "New York - NY",
    "Namespace": "namespace1"
  },
  {
    "ID": "1_Mary",
    "CreatedAt": "2022-05-18T11:49:22.278075579-03:00",
    "ParentName": "John",
    "ParentBirthYear": 1992,
    "ChildName": "Mary",
    "CityAndState": "New York - NY",
    "Namespace": "namespace1"
  },
  {
    "ID": "3_Bob",
    "CreatedAt": "2022-05-18T11:49:22.278036367-03:00",
    "ParentName": "Chris",
    "ParentBirthYear": 1987,
    "ChildName": "Bob",
    "CityAndState": "Washington - DC",
    "Namespace": "namespace1"
  },
  {
    "ID": "1_John",
    "CreatedAt": "2022-05-18T11:49:22.278003329-03:00",
    "ParentName": "John",
    "ParentBirthYear": 1992,
    "ChildName": "John",
    "CityAndState": "New York - NY",
    "Namespace": "namespace1"
  },
  {
    "ID": "3_Bob",
    "CreatedAt": "2022-05-18T11:49:22.27810943-03:00",
    "ParentName": "Chris",
    "ParentBirthYear": 1987,
    "ChildName": "Bob",
    "CityAndState": "Washington - DC",
    "Namespace": "namespace2"
  }
]
 <nil>
*/
