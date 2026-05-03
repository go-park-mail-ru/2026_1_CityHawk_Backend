package grpcconv

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TimeToProto(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t.UTC())
}

func OptionalTimeToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return TimeToProto(*t)
}

func TimeFromProto(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	value := t.AsTime().UTC()
	return &value
}
