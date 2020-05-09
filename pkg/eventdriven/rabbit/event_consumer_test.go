package rabbit

import "github.com/ztimes2/jazzba/pkg/eventdriven"

var (
	_ eventdriven.Consumer = (*EventConsumer)(nil)
)

// TODO: add tests
