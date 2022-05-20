# GoMsgProcessor

### A Golang library for parallel processing of messages to structured documents

---------------------

## Table of Contents

  - [1. Description](#Description)
  - [2. Technology Stack](#TechnologyStack)
  - [3. Getting Started](#GettingStarted)
  - [4. Changelog](#Changelog)
  - [5. Collaborators](#Collaborators)
  - [6. Contributing](#Contributing)
  - [7. License](#License)
  - [8. Contact Information](#ContactInformation)

## <a name="Description" /> 1. Description

GoMsgProcessor is a generic library to read messages, in a recursively and parallel way, requiring only a builder to transform then to final documents. It is possible to set multiple builders, associating each one with a message type, allowing to work with messages from different sources. Through the namespaces, it is also be able to work with different targets. In addition, a deduplication function can be injected to clean up the slice of documents after the process.

## <a name="TechnologyStack" /> 2. Technology Stack

| **Stack**     | **Version** |
|---------------|-------------|
| Golang        | v1.18       |
| golangci-lint | v1.46.2     |

## <a name="GettingStarted" /> 3. Getting Started

- ### <a name="Prerequisites" /> Prerequisites

  - Any [Golang](https://go.dev/doc/install) programming language version installed, preferred 1.18 or later.

- ### <a name="Install" /> Install
  
  ```
  go get -u github.com/arquivei/gomsgprocessor
  ```

- ### <a name="ConfigurationSetup" /> Configuration Setup

  ```
  go mod vendor
  go mod tidy
  ```

- ### <a name="Usage" /> Usage
  
  - Import the package

    ```go
    import (
        "github.com/arquivei/gomsgprocessor"
    )
    ```

  - Define a incoming message struct

    ```go
    type ExampleMessage struct {
        ID            int      `json:"id"`
        Name          string   `json:"name"`
        Age           int      `json:"age"`
        City          string   `json:"city"`
        State         string   `json:"state"`
        ChildrenNames []string `json:"childrenNames"`
        Namespace     string   `json:"namespace"`
    }
    ```

  - Implement the Message interface, witch is the input of ParallelProcessor's MakeDocuments.

    ```go
    func (e *ExampleMessage) GetNamespace() gomsgprocessor.Namespace {
        // Namespace is a logical separator that will be used to group messages while
		// processing then.
        return gomsgprocessor.Namespace(e.Namespace)
    }

    func (e *ExampleMessage) GetType() gomsgprocessor.MessageType {
        // MessageType is used to decide which DocumentBuilder to use for each Message.
        return gomsgprocessor.MessageType("typeExample")
    }

    func (e *ExampleMessage) UpdateLogWithData(ctx context.Context) {
        // Optional logger method
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
    ```

  - Define a outcoming document struct, witch is the result of a DocumentBuilder's Build.

    ```go
    type ExampleDocument struct {
        ID              string
        CreatedAt       time.Time
        ParentName      string
        ParentBirthYear int
        ChildName       string
        CityAndState    string
        Namespace       string
    }
    ```

  - Implement the DocumentBuilder interface, witch transforms a Message into a slice of Documents.

    ```go
    type ExampleBuilder struct{}

    // Build transforms a Message into []Document.
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
    ```
  
  - Define a (optional) function, used for deduplicate the slice of documents. 

    ```go
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
    ```

  - And now, it's time!

    ```go
    func main() {

		// NewParallelProcessor returns a new ParallelProcessor with a map of
		// DocumentBuilder for each MessageType.
		//
		// A list of Option is also available for this method. See option.go for more
		// information.
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

		// MakeDocuments creates in parallel a slice of Document for given []Message
		// using the map of DocumentBuilder (see NewParallelProcessor).
		//
		// This method returns a []Document and a (foundationkit/errors).Error.
		// If not nil, this error has a (foundationkit/errors).Code associated with and
		// can be a ErrCodeBuildDocuments or a ErrCodeDeduplicateDocuments.
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

	// Simple json marshaler with indentation
	func JSONMarshal(t interface{}) (string, error) {
		buffer := &bytes.Buffer{}
		encoder := json.NewEncoder(buffer)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		err := encoder.Encode(t)
		return buffer.String(), err
	}
    ```

- ### <a name="Examples" /> Examples
  
  - [Sample usage](https://github.com/arquivei/gomsgprocessor/blob/master/examples/main.go)

## <a name="Changelog" /> 4. Changelog

  - **GoMsgProcessor 0.1.0 (May 20, 2022)**
  
    - [New] Decoupling this package from Arquivei's API projects.
    - [New] Setting github's workflow with golangci-lint 
    - [New] Example for usage.
    - [New] Documents: Code of Conduct, Contributing, License and Readme.

## <a name="Collaborators" /> 5. Collaborators

- ### <a name="Authors" /> Authors
  
  <!-- markdownlint-disable -->
  <!-- prettier-ignore-start -->
	<table>
	<tr>
		<td align="center"><a href="https://github.com/victormn"><img src="https://avatars.githubusercontent.com/u/9757545?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Victor Nunes</b></sub></a></td>
		<td align="center"><a href="https://github.com/rjfonseca"><img src="https://avatars.githubusercontent.com/u/151265?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Rodrigo Fonseca</b></sub></a></td>
	</tr>
	</table>
  <!-- markdownlint-restore -->
  <!-- prettier-ignore-end -->

- ### <a name="Maintainers" /> Maintainers
  
  <!-- markdownlint-disable -->
  <!-- prettier-ignore-start -->
	<table>
	<tr>
		<td align="center"><a href="https://github.com/rilder-almeida"><img src="https://avatars.githubusercontent.com/u/49083200?v=4s=100" width="100px;" alt=""/><br /><sub><b>Rilder Almeida</b></sub></a></td>
	</tr>
	</table>
  <!-- markdownlint-restore -->
  <!-- prettier-ignore-end -->

## <a name="Contributing" /> 6. Contributing

  Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## <a name="License" /> 7. License
  
  We use [Semantic Versioning](http://semver.org/) for versioning. For the versions
  available, see the [tags on this repository](https://github.com/arquivei/gomsgprocessor/tags).

## <a name="ContactInformation" /> 8. Contact Information

  All contact may be doing by [rilder.almeida@arquivei.com.br](mailto:rilder.almeida@arquivei.com.br)
