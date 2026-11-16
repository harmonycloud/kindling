// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package internal

import "testing"

func TestCallback(t *testing.T) {
	connMap := make(map[ConnKey]*ConnectionStats)
	connKey := ConnKey{
		SrcIP:   "10.10.10.10",
		SrcPort: 40040,
		DstIP:   "10.10.10.23",
		DstPort: 80,
	}
	statesResource := createStatesResource()
	connStats := &ConnectionStats{
		Pid:              0,
		Comm:             "test",
		ConnKey:          connKey,
		InitialTimestamp: 0,
		EndTimestamp:     0,
		Code:             0,
	}
	connStats.StateMachine = NewStateMachine(Inprogress, statesResource, connStats)
	connMap[connKey] = connStats

	stats, err := connStats.StateMachine.ReceiveEvent(tcpConnectNoError, connMap)
	if err != nil {
		t.Fatal(err)
	}
	if connStats.StateMachine.currentStateType != Inprogress {
		t.Errorf("expected inprogress, got %v", connStats.StateMachine.currentStateType)
	}

	stats, err = connStats.StateMachine.ReceiveEvent(tcpSetStateToEstablished, connMap)
	if err != nil {
		t.Fatal(err)
	}
	if stats == nil {
		t.Errorf("expected stats emitted, but none got")
	}
	if connStats.StateMachine.currentStateType != Success {
		t.Errorf("expected success, got %v", connStats.StateMachine.currentStateType)
	}
}
