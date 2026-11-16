// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"go.opentelemetry.io/otel/attribute"
)

// Adapter is used to transform *model.DataGroup into Trace or Metric
type Adapter interface {
	Adapt(group *model.DataGroup, attrType AttrType) ([]*AdaptedResult, error)
}

type AdaptedResult struct {
	ResultType ResultType

	// AttrsList contains labels for sync Metric and Trace
	AttrsList []attribute.KeyValue
	// AttrsMap contains labels for Async Metric
	AttrsMap  *model.AttributeMap
	Metrics   []*model.Metric
	Timestamp uint64

	// FreeAttrsMap provides an interface for those adapters which need to reuse AttrsList and AttrsMap
	FreeAttrsMap
	FreeAttrsList
}

type AttrType string

type FreeAttrsMap func(attrsMap *model.AttributeMap)
type FreeAttrsList func(attrsList []attribute.KeyValue)

type ResultType string

const (
	Metric ResultType = "metric"
	Trace  ResultType = "trace"
)

const (
	AttributeList AttrType = "attributeList"
	AttributeMap  AttrType = "attributeMap"
)

func (r *AdaptedResult) Free() {
	if r.FreeAttrsMap != nil {
		r.FreeAttrsMap(r.AttrsMap)
	}
	if r.FreeAttrsList != nil {
		r.FreeAttrsList(r.AttrsList)
	}
}
