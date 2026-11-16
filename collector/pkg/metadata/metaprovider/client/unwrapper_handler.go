// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package metadataclient

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"
	"github.com/Kindling-project/kindling/collector/pkg/metadata/metaprovider/api"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type unwrapper func([]byte) (interface{}, error)

// globalOption to show each message received from MP
var enableTrace bool

func SetEnableTraceFromMPClient(enable bool) {
	enableTrace = enable
}

type UnwrapperHandler struct {
	add    api.AddObj
	update api.UpdateObj
	delete api.DeleteObj

	unwrapper
}

var podUnwrapperHander = NewUnwrapperHander(
	kubernetes.AddPod,
	kubernetes.UpdatePod,
	kubernetes.DeletePod,
	func(b []byte) (interface{}, error) {
		obj := corev1.Pod{}
		err := json.Unmarshal(b, &obj)
		return &obj, err
	},
)

var serviceUnwrapperHander = NewUnwrapperHander(
	kubernetes.AddService,
	kubernetes.UpdateService,
	kubernetes.DeleteService,
	func(b []byte) (interface{}, error) {
		obj := corev1.Service{}
		err := json.Unmarshal(b, &obj)
		return &obj, err
	},
)

var nodeUnwrapperHander = NewUnwrapperHander(
	kubernetes.AddNode,
	kubernetes.UpdateNode,
	kubernetes.DeleteNode,
	func(b []byte) (interface{}, error) {
		obj := corev1.Node{}
		err := json.Unmarshal(b, &obj)
		return &obj, err
	},
)

var relicaSetUnwrapperHander = NewUnwrapperHander(
	kubernetes.AddReplicaSet,
	kubernetes.UpdateReplicaSet,
	kubernetes.DeleteReplicaSet,
	func(b []byte) (interface{}, error) {
		obj := appv1.ReplicaSet{}
		err := json.Unmarshal(b, &obj)
		return &obj, err
	},
)

func NewUnwrapperHander(add api.AddObj, update api.UpdateObj, delete api.DeleteObj, unwrapper unwrapper) UnwrapperHandler {
	return UnwrapperHandler{
		add:       add,
		update:    update,
		delete:    delete,
		unwrapper: unwrapper,
	}
}

func (uw *UnwrapperHandler) Apply(resp *api.MetaDataVO) error {
	if enableTrace {
		log.Println(api.MetaDataVO2String(resp))
	}
	switch resp.Operation {
	case string(api.Add):
		if resp.NewObj == nil {
			return fmt.Errorf("operation [add] missing data [newObj]")
		}
		if obj, err := uw.unwrapper(resp.NewObj); err == nil {
			uw.add(obj)
			return nil
		} else {
			return err
		}
	case string(api.Update):
		if resp.NewObj == nil {
			return fmt.Errorf("operation [update] missing data [newObj]")
		}
		if resp.OldObj == nil {
			return fmt.Errorf("operation [update] missing data [oldObj]")
		}
		oldObj, err := uw.unwrapper(resp.OldObj)
		if err != nil {
			return err
		}
		newObj, err := uw.unwrapper(resp.NewObj)
		if err != nil {
			return err
		}
		uw.update(oldObj, newObj)
		return nil
	case string(api.Delete):
		if resp.OldObj == nil {
			return fmt.Errorf("operation [delete] missing data [oldObj]")
		}
		if obj, err := uw.unwrapper(resp.OldObj); err == nil {
			uw.delete(obj)
			return nil
		} else {
			return err
		}
	default:
		return fmt.Errorf("unexpect operation: %s", resp.Operation)
	}
}
