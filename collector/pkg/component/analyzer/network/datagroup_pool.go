// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"sync"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
)

func createDataGroup() interface{} {
	values := []*model.Metric{
		model.NewIntMetric(constvalues.ConnectTime, 0),
		model.NewIntMetric(constvalues.RequestSentTime, 0),
		model.NewIntMetric(constvalues.WaitingTtfbTime, 0),
		model.NewIntMetric(constvalues.ContentDownloadTime, 0),
		model.NewIntMetric(constvalues.RequestTotalTime, 0),
		model.NewIntMetric(constvalues.RequestIo, 0),
		model.NewIntMetric(constvalues.ResponseIo, 0),
	}
	dataGroup := model.NewDataGroup(constnames.NetRequestMetricGroupName, model.NewAttributeMap(), uint64(time.Now().UnixNano()), values...)
	return dataGroup
}

type DataGroupPool interface {
	Get() *model.DataGroup
	Free(dataGroup *model.DataGroup)
}

type SimpleDataGroupPool struct {
	pool *sync.Pool
}

func NewDataGroupPool() DataGroupPool {
	return &SimpleDataGroupPool{pool: &sync.Pool{New: createDataGroup}}
}

func (p *SimpleDataGroupPool) Get() *model.DataGroup {
	return p.pool.Get().(*model.DataGroup)
}

func (p *SimpleDataGroupPool) Free(dataGroup *model.DataGroup) {
	dataGroup.Reset()
	dataGroup.Name = constnames.NetRequestMetricGroupName
	p.pool.Put(dataGroup)
}
