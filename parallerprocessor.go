package gomsgprocessor

import (
	"context"

	"github.com/arquivei/foundationkit/errors"
	"github.com/arquivei/foundationkit/trace"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// ParallelProcessor is a interface that process in parallel a slice of Message.
type ParallelProcessor interface {
	MakeDocuments(context.Context, []Message) ([]Document, error)
}

// DocumentBuilder is a interface that transforms a Message into []Document.
type DocumentBuilder interface {
	Build(context.Context, Message) ([]Document, error)
}

type parallelProcessor struct {
	builders             map[MessageType]DocumentBuilder
	deduplicateDocuments DeduplicateDocumentsFunc
}

// NewParallelProcessor returns a new ParallelProcessor with a map of
// DocumentBuilder for each MessageType.
//
// A list of Option is also available for this method. See option.go for more
// information.
func NewParallelProcessor(
	builders map[MessageType]DocumentBuilder,
	opts ...Option,
) ParallelProcessor {
	p := &parallelProcessor{
		builders:             builders,
		deduplicateDocuments: defaultDeduplicateDocumentsFunc,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// MakeDocuments creates in parallel a slice of Document for given []Message
// using the map of DocumentBuilder (see NewParallelProcessor).
//
// This method returns a []Document and a (foundationkit/errors).Error.
// If not nil, this error has a (foundationkit/errors).Code associated with and
// can be a ErrCodeBuildDocuments or a ErrCodeDeduplicateDocuments.
func (p *parallelProcessor) MakeDocuments(ctx context.Context, msgs []Message) ([]Document, error) {
	const op = errors.Op("gomsgprocessor.parallelProcessor.MakeDocuments")

	ctx, span := trace.StartSpan(ctx, op.String())
	defer span.End(nil)

	documentsByNamespace, err := p.parallelBuildDocumentsByNamespace(ctx, msgs)
	if err != nil {
		return nil, errors.E(op, err, ErrCodeBuildDocuments)
	}

	deduplicatedDocuments, err := p.deduplicateDocumentsForEachNamespace(documentsByNamespace)
	if err != nil {
		return nil, errors.E(op, err, ErrCodeDeduplicateDocuments)
	}

	return deduplicatedDocuments, nil
}

func (p *parallelProcessor) parallelBuildDocumentsByNamespace(
	ctx context.Context,
	msgs []Message,
) (map[Namespace][]Document, error) {
	const op = errors.Op("parallelBuildDocumentsByNamespace")

	type builtDocument struct {
		documents []Document
		namespace Namespace
	}

	builtDocuments := make([]builtDocument, len(msgs))

	g, ctx := errgroup.WithContext(ctx)
	for i, msg := range msgs {
		i, msg := i, msg

		g.Go(func() error {
			documentBuilder, ok := p.builders[msg.GetType()]
			if !ok {
				msg.UpdateLogWithData(ctx)
				return errors.E(ErrMsgTypeHasNoBuilder, errors.KV("type", msg.GetType()))
			}

			documents, err := documentBuilder.Build(ctx, msg)
			if err != nil {
				msg.UpdateLogWithData(ctx)
				return err
			}

			if documents == nil {
				msg.UpdateLogWithData(ctx)
				log.Ctx(ctx).Info().Msg("Message ignored...")
				return nil
			}

			builtDocuments[i] = builtDocument{
				documents: documents,
				namespace: msg.GetNamespace(),
			}

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return nil, errors.E(op, err)
	}

	documentsByNamespace := make(map[Namespace][]Document, len(builtDocuments))
	for _, builtDoc := range builtDocuments {
		if builtDoc.documents == nil || builtDoc.namespace == "" {
			continue
		}
		documentsByNamespace[builtDoc.namespace] = append(
			documentsByNamespace[builtDoc.namespace],
			builtDoc.documents...,
		)
	}
	return documentsByNamespace, nil
}

func (p *parallelProcessor) deduplicateDocumentsForEachNamespace(
	documentsByNamespace map[Namespace][]Document,
) ([]Document, error) {
	const op = errors.Op("deduplicateDocumentsForEachNamespace")

	documents := make([]Document, 0, len(documentsByNamespace))
	for _, docs := range documentsByNamespace {
		deduplicatedDocuments, err := p.deduplicateDocuments(docs)
		if err != nil {
			return nil, errors.E(op, err)
		}
		documents = append(documents, deduplicatedDocuments...)
	}
	return documents, nil
}
