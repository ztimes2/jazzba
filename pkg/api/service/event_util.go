package service

import (
	"github.com/ztimes2/jazzba/pkg/eventdriven"

	"github.com/pkg/errors"
)

func newEventProducingError(eventType eventdriven.EventType, err error) error {
	return errors.Wrapf(err, "could not produce event '%s'", eventType)
}
