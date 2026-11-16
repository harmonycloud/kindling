// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
)

type StateType string

type StateMachine struct {
	connStats        *ConnectionStats
	lastStateType    StateType
	currentStateType StateType
	states           StatesResource
}

func NewStateMachine(initialState StateType, statesResource StatesResource, connStats *ConnectionStats) *StateMachine {
	return &StateMachine{
		connStats:        connStats,
		lastStateType:    initialState,
		currentStateType: initialState,
		states:           statesResource,
	}
}

func (s *StateMachine) GetCurrentState() StateType {
	return s.currentStateType
}

func (s *StateMachine) GetLastState() StateType {
	return s.lastStateType
}

func (s *StateMachine) ReceiveEvent(event EventType, connMap map[ConnKey]*ConnectionStats) (*ConnectionStats, error) {
	currentState, ok := s.states[s.currentStateType]
	if !ok {
		return nil, fmt.Errorf("no current state [%v]", s.currentStateType)
	}
	nextStateType, ok := currentState.eventsMap[event]
	if !ok {
		return nil, fmt.Errorf("receive not supported event [%v] in state [%v], ConnKey: [%s]",
			event, s.currentStateType, s.connStats.ConnKey.String())
	}
	nextState, ok := s.states[nextStateType]
	if !ok {
		return nil, fmt.Errorf("no next state [%v]", nextStateType)
	}
	// Only trigger the callback when the state changes
	if nextStateType != s.currentStateType && nextState.callback != nil {
		s.lastStateType = s.currentStateType
		s.currentStateType = nextStateType
		return nextState.callback(s.connStats, connMap), nil
	}
	s.lastStateType = s.currentStateType
	s.currentStateType = nextStateType
	return nil, nil
}

type Callback func(connStats *ConnectionStats, connMap map[ConnKey]*ConnectionStats) *ConnectionStats

type EventType string

type State struct {
	eventsMap map[EventType]StateType
	callback  Callback
}

type StatesResource map[StateType]State
