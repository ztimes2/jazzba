package rabbit

import "github.com/ztimes2/jazzba/pkg/eventdriven"

var (
	_ eventdriven.Producer = (*EventProducer)(nil)
)

// TODO: add tests
