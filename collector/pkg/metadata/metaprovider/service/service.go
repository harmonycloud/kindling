// Copyright (C) 2023-2024 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"
	"github.com/Kindling-project/kindling/collector/pkg/metadata/metaprovider/api"
	"github.com/Kindling-project/kindling/collector/pkg/metadata/metaprovider/ioutil"
)

var sem = semaphore.NewWeighted(int64(16))

type MetaDataWrapper struct {
	flushersMap                  sync.Map
	historyClientCount           atomic.Int32
	historyDisconnectClientCount atomic.Int32
	// Signal
	stopCh chan struct{}

	ReceivedPodsEventCount    EventCount
	ReceivedRsEventCount      EventCount
	ReceivedServiceEventCount EventCount
	ReceivedNodesEventCount   EventCount
}

type EventCount struct {
	Add    atomic.Int64
	Delete atomic.Int64
	Update atomic.Int64
}

func (c *EventCount) AddEvent(operation string) {
	switch operation {
	case "add":
		c.Add.Add(1)
	case "update":
		c.Update.Add(1)
	case "delete":
		c.Delete.Add(1)
	}
}

func (c *EventCount) String() string {
	return fmt.Sprintf("[ Add: %d,Update: %d,Delete: %d ]", c.Add.Load(), c.Update.Load(), c.Delete.Load())
}

func NewMetaDataWrapper(config *Config) (*MetaDataWrapper, error) {
	mp := &MetaDataWrapper{
		stopCh: make(chan struct{}),
	}

	// DEBUG Watcher Size
	if config.LogInterval > 0 {
		go func() {
			log.Printf("Log Status interval: %d(s)", config.LogInterval)
			ticker := time.NewTicker(time.Duration(config.LogInterval) * time.Second)
			for {
				select {
				case <-ticker.C:
					log.Printf("Received Events: Pod: %s; ReplicaSet: %s; Node: %s; Service: %s\n",
						mp.ReceivedPodsEventCount.String(),
						mp.ReceivedRsEventCount.String(),
						mp.ReceivedNodesEventCount.String(),
						mp.ReceivedServiceEventCount.String(),
					)

					log.Printf("Cached Resources Counts: Containers:%d, ReplicaSet: %d\n",
						kubernetes.MetaDataCache.GetCacheContainerIdInfoSize(),
						kubernetes.GlobalRsInfo.GetSize(),
					) // kubernetes.GlobalPodInfo.Info
				case <-mp.stopCh:
					ticker.Stop()
					return
				}
			}
		}()
	}

	var options []kubernetes.Option
	options = append(options, kubernetes.WithAuthType(config.KubeAuthType))
	options = append(options, kubernetes.WithKubeConfigDir(config.KubeConfigDir))
	options = append(options, kubernetes.WithGraceDeletePeriod(0))
	options = append(options, kubernetes.WithFetchReplicaSet(config.EnableFetchReplicaSet))
	options = append(options, kubernetes.WithPodEventHander(NewHandler(
		"pod",
		kubernetes.AddPod,
		kubernetes.UpdatePod,
		kubernetes.DeletePod,
		mp.broadcast,
	)))
	options = append(options, kubernetes.WithReplicaSetEventHander(NewHandler(
		"rs",
		kubernetes.AddReplicaSet,
		kubernetes.UpdateReplicaSet,
		kubernetes.DeleteReplicaSet,
		mp.broadcast,
	)))
	options = append(options, kubernetes.WithNodeEventHander(NewHandler(
		"node",
		kubernetes.AddNode,
		kubernetes.UpdateNode,
		kubernetes.DeleteNode,
		mp.broadcast,
	)))
	options = append(options, kubernetes.WithServiceEventHander(NewHandler(
		"service",
		kubernetes.AddService,
		kubernetes.UpdateService,
		kubernetes.DeleteService,
		mp.broadcast,
	)))
	return mp, kubernetes.InitK8sHandler(options...)
}

func (s *MetaDataWrapper) broadcast(data api.MetaDataRsyncResponse) {
	switch data.Type {
	case "pod":
		s.ReceivedPodsEventCount.AddEvent(data.Operation)
	case "rs":
		s.ReceivedRsEventCount.AddEvent(data.Operation)
	case "node":
		s.ReceivedNodesEventCount.AddEvent(data.Operation)
	case "service":
		s.ReceivedServiceEventCount.AddEvent(data.Operation)
	}
	b, _ := json.Marshal(data)
	idleMax := 10 * time.Second
	idleTimeout := time.NewTimer(idleMax)
	defer idleTimeout.Stop()
	s.flushersMap.Range(func(_, flusher interface{}) bool {
		idleTimeout.Reset(idleMax)
		f := flusher.(*Watcher)
		// 如果一个事件5s还未正常写入到发送至目标节点(有10个事件的缓冲区),则认为目标客户端不可用，关闭写入通道，并要求结束
		select {
		case f.eventChannel <- append(b, '\n'):
		case <-idleTimeout.C:
			log.Printf("Event Flush Timeout, Disconnect client [%s] from server side!\n", f.IP)
			f.Close()
			s.RemoveFlusher(f)
			return true
		}
		return true
	})
}

func (s *MetaDataWrapper) list() ([]byte, error) {
	resp := api.ListVO{
		Cache:             kubernetes.MetaDataCache,
		GlobalNodeInfo:    kubernetes.GlobalNodeInfo,
		GlobalRsInfo:      kubernetes.GlobalRsInfo,
		GlobalServiceInfo: kubernetes.GlobalServiceInfo,
	}
	return json.Marshal(resp)
}

func GetIPFromRequest(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Real-IP")
	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	ip = r.Header.Get("X-Forward-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	return "", errors.New("no valid ip found")
}

func (s *MetaDataWrapper) ListAndWatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	f := ioutil.NewWriteFlusher(w)
	ip, err := GetIPFromRequest(r)
	if err != nil {
		ip = "unknow"
	}
	if watcher, err := s.ListWithSemaphore(ctx, f, ip); err != nil {
		log.Printf("Failed to acquire semaphore: %v", err)
		w.WriteHeader(500)
		return
	} else {
		watcher.watch(ctx, s.stopCh)
		s.RemoveFlusher(watcher)
	}
}

func (s *MetaDataWrapper) ListWithSemaphore(ctx context.Context, f *ioutil.WriteFlusher, ip string) (*Watcher, error) {
	if err := sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer sem.Release(1)

	kubernetes.RLockMetadataCache()
	defer kubernetes.RUnlockMetadataCache()
	b, _ := s.list()
	f.Write(append(b, '\n'))

	w := &Watcher{
		eventChannel: make(chan []byte, 10),
		IP:           ip,
		WriteFlusher: f,
	}
	s.AddFlusher(w)
	addCount := s.historyClientCount.Load()
	disconnectCount := s.historyDisconnectClientCount.Load()
	log.Printf("Add Watcher , client is from IP: %s , current Watcher Size : %d", ip, addCount-disconnectCount)
	return w, nil
}

type Watcher struct {
	Id           int32
	IP           string
	eventChannel chan []byte
	*ioutil.WriteFlusher

	isClosed atomic.Bool
}

func (s *MetaDataWrapper) AddFlusher(w *Watcher) {
	w.Id = s.historyClientCount.Add(1)
	w.isClosed.Store(false)
	s.flushersMap.Store(w.Id, w)
}

func (s *MetaDataWrapper) RemoveFlusher(w *Watcher) {
	if w.isClosed.Swap(true) {
		// 如果旧值为 true，说明已经被boardcast线程关闭过,直接退出
		return
	}
	disconnectCount := s.historyDisconnectClientCount.Add(1)
	addCount := s.historyClientCount.Load()
	log.Printf("Remove Watcher: client from IP: %s is disconnected, current Watcher Size : %d", w.IP, addCount-disconnectCount)
	s.flushersMap.Delete(w.Id)
}

func (w *Watcher) watch(ctx context.Context, stopCh <-chan struct{}) {
	defer w.Close()
	for {
		select {
		case data := <-w.eventChannel:
			_, err := w.WriteFlusher.Write(data)
			if err != nil {
				log.Printf("remote Connection closed durning flush, Error: %v\n", err)
				return
			}
		case <-ctx.Done():
			// 客户端正常断开
			return
		case <-stopCh:
			// TODO clear eventChannel and return
			return
		}
	}
}

func (s *MetaDataWrapper) Shutdown() {
	// TODO Check channel
	close(s.stopCh)
}
