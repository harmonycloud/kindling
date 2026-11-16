// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/model"
)

// ConnectMonitor reads in events related to TCP connect operations and updates its
// status to record the connection procedure.
// This is not thread safe to use.
type ConnectMonitor struct {
	connMap        map[ConnKey]*ConnectionStats
	statesResource StatesResource
	hostProcPath   string
	logger         *component.TelemetryLogger
}

const HostProc = "HOST_PROC_PATH"

func NewConnectMonitor(logger *component.TelemetryLogger) *ConnectMonitor {
	path, ok := os.LookupEnv(HostProc)
	if !ok {
		path = "/proc"
	}
	return &ConnectMonitor{
		connMap:        make(map[ConnKey]*ConnectionStats),
		statesResource: createStatesResource(),
		hostProcPath:   path,
		logger:         logger,
	}
}

func (c *ConnectMonitor) ReadInConnectExitSyscall(event *model.KindlingEvent) (*ConnectionStats, error) {
	retValue := event.GetUserAttribute("res")
	if retValue == nil {
		return nil, fmt.Errorf("res of connect_exit is nil")
	}
	retValueInt := retValue.GetIntValue()

	connKey := ConnKey{
		SrcIP:   event.GetSip(),
		SrcPort: event.GetSport(),
		DstIP:   event.GetDip(),
		DstPort: event.GetDport(),
	}
	if ce := c.logger.Check(zapcore.DebugLevel, "Receive connect_exit event:"); ce != nil {
		ce.Write(
			zap.String("ConnKey", connKey.String()),
			zap.Int64("retValue", retValueInt),
		)
	}

	connStats, ok := c.connMap[connKey]
	if !ok {
		// Maybe the connStats have been closed by tcp_set_state_from_established event.
		// We don't care about it.
		return nil, nil
	}
	// "connect_exit" comes to analyzer after "tcp_connect"
	connStats.EndTimestamp = event.Timestamp
	connStats.Pid = event.GetPid()
	connStats.Comm = event.GetComm()
	connStats.ContainerId = event.GetContainerId()
	var eventType EventType
	if retValueInt == 0 {
		eventType = connectExitSyscallSuccess
	} else if isNotErrorReturnCode(retValueInt) {
		eventType = connectExitSyscallNotConcern
	} else {
		eventType = connectExitSyscallFailure
		connStats.Code = int(retValueInt)
	}
	return connStats.StateMachine.ReceiveEvent(eventType, c.connMap)
}

func (c *ConnectMonitor) ReadSendRequestSyscall(event *model.KindlingEvent) (*ConnectionStats, error) {
	// The events without sip/sport/dip/dport have been filtered outside this method.
	connKey := ConnKey{
		SrcIP:   event.GetSip(),
		SrcPort: event.GetSport(),
		DstIP:   event.GetDip(),
		DstPort: event.GetDport(),
	}
	if ce := c.logger.Check(zapcore.DebugLevel, "Receive sendRequestSyscall event:"); ce != nil {
		ce.Write(
			zap.String("ConnKey", connKey.String()),
			zap.String("eventName", event.Name),
		)
	}

	connStats, ok := c.connMap[connKey]
	if !ok {
		return nil, nil
	}
	connStats.EndTimestamp = event.Timestamp
	connStats.Pid = event.GetPid()
	connStats.Comm = event.GetComm()
	connStats.ContainerId = event.GetContainerId()
	return connStats.StateMachine.ReceiveEvent(sendRequestSyscall, c.connMap)
}

func isNotErrorReturnCode(code int64) bool {
	return code == einprogress || code == eintr || code == eisconn || code == ealready
}

func (c *ConnectMonitor) ReadInTcpConnect(event *model.KindlingEvent) (*ConnectionStats, error) {
	connKey, err := getConnKeyForTcpConnect(event)
	if err != nil {
		return nil, err
	}
	retValue := event.GetUserAttribute("retval")
	if retValue == nil {
		return nil, fmt.Errorf("retval of tcp_connect is nil")
	}
	retValueInt := retValue.GetUintValue()

	if ce := c.logger.Check(zapcore.DebugLevel, "Receive tcp_connect event:"); ce != nil {
		ce.Write(
			zap.String("ConnKey", connKey.String()),
			zap.Uint64("retValue", retValueInt),
		)
	}

	var eventType EventType
	if retValueInt == 0 {
		eventType = tcpConnectNoError
	} else {
		eventType = tcpConnectError
	}

	connStats, ok := c.connMap[connKey]
	if !ok {
		// "tcp_connect" comes to analyzer before "connect_exit"
		connStats = &ConnectionStats{
			ConnKey:          connKey,
			InitialTimestamp: event.Timestamp,
			EndTimestamp:     event.Timestamp,
			Code:             int(retValueInt),
		}
		connStats.StateMachine = NewStateMachine(Inprogress, c.statesResource, connStats)
		c.connMap[connKey] = connStats
	} else {
		// Not possible to enter this branch
		c.logger.Info("Receive another unexpected tcp_connect event", zap.String("connKey", connKey.String()))
		connStats.EndTimestamp = event.Timestamp
		connStats.Code = int(retValueInt)
	}
	return connStats.StateMachine.ReceiveEvent(eventType, c.connMap)
}

const (
	establishedState = 1
)

func (c *ConnectMonitor) ReadInTcpSetState(event *model.KindlingEvent) (*ConnectionStats, error) {
	connKey, err := getConnKeyForTcpConnect(event)
	if err != nil {
		return nil, err
	}

	oldState := event.GetUserAttribute("old_state")
	newState := event.GetUserAttribute("new_state")
	if oldState == nil || newState == nil {
		return nil, fmt.Errorf("tcp_set_state events have nil state, ConnKey: %s", connKey.String())
	}
	oldStateInt := oldState.GetIntValue()
	newStateInt := newState.GetIntValue()

	if oldStateInt == establishedState {
		return c.readInTcpSetStateFromEstablished(connKey, event)
	} else if newStateInt == establishedState {
		return c.readInTcpSetStateToEstablished(connKey, event)
	} else {
		return nil, fmt.Errorf("no state is 'established' for tcp_set_state event, "+
			"old state: %d, new state: %d", oldStateInt, newStateInt)
	}
}

func (c *ConnectMonitor) readInTcpSetStateToEstablished(connKey ConnKey, event *model.KindlingEvent) (*ConnectionStats, error) {
	if ce := c.logger.Check(zapcore.DebugLevel, "Receive tcp_set_state(to established) event:"); ce != nil {
		ce.Write(
			zap.String("ConnKey", connKey.String()),
		)
	}
	connStats, ok := c.connMap[connKey]
	if !ok {
		// No tcp_connect or connect_exit received.
		// This is the events from server-side.
		c.logger.Debug("No tcp_connect received, but receive tcp_set_state_to_established")
		return nil, nil
	}
	connStats.EndTimestamp = event.Timestamp
	return connStats.StateMachine.ReceiveEvent(tcpSetStateToEstablished, c.connMap)
}

func (c *ConnectMonitor) readInTcpSetStateFromEstablished(connKey ConnKey, event *model.KindlingEvent) (*ConnectionStats, error) {
	if ce := c.logger.Check(zapcore.DebugLevel, "Receive tcp_set_state(from established) event:"); ce != nil {
		ce.Write(
			zap.String("ConnKey", connKey.String()),
		)
	}
	connStats, ok := c.connMap[connKey]
	if !ok {
		// Connection has been established and the connStats have been emitted.
		return nil, nil
	}
	connStats.EndTimestamp = event.Timestamp
	return connStats.StateMachine.ReceiveEvent(tcpSetStateFromEstablished, c.connMap)
}

func (c *ConnectMonitor) TrimConnectionsWithTcpStat(waitForEventSecond int) []*ConnectionStats {
	ret := make([]*ConnectionStats, 0, len(c.connMap))
	// Only scan once for each Pid
	pidTcpStateMap := make(map[uint32]NetSocketStateMap)
	waitForEventNano := int64(waitForEventSecond) * 1000000000
	timeNow := time.Now().UnixNano()
	for key, connStat := range c.connMap {
		if connStat.Pid == 0 {
			continue
		}
		if timeNow-int64(connStat.InitialTimestamp) < waitForEventNano {
			// Still waiting for other events
			continue
		}
		tcpStateMap, ok := pidTcpStateMap[connStat.Pid]
		if !ok {
			tcpState, err := NewPidTcpStat(c.hostProcPath, int(connStat.Pid))
			if err != nil {
				c.logger.Debug("error happened when scanning net/tcp",
					zap.Uint32("Pid", connStat.Pid), zap.Error(err))
				// No such file or directory, which means the process has been purged.
				// We consider the connection failed to be established.
				stats, err := connStat.StateMachine.ReceiveEvent(expiredEvent, c.connMap)
				if err != nil {
					c.logger.Warn("error happened when receiving event:", zap.Error(err))
				}
				if stats != nil {
					ret = append(ret, stats)
				}
				continue
			}
			pidTcpStateMap[connStat.Pid] = tcpState
			tcpStateMap = tcpState
		}
		state, ok := tcpStateMap[key.toSocketKey()]
		// There are two possible reasons for no such socket found:
		//   1. The connection was established and has been closed.
		//      There should have been received a tcp_set_state event, and c.connMap
		//      should not contain such socket. In this case, we won't enter this piece of code.
		//   2. The connection failed to be established and has been closed.
		if !ok {
			stats, err := connStat.StateMachine.ReceiveEvent(expiredEvent, c.connMap)
			if err != nil {
				c.logger.Warn("error happened when receiving event:", zap.Error(err))
			}
			if stats != nil {
				ret = append(ret, stats)
			}
			continue
		}
		if state == established {
			stats, err := connStat.StateMachine.ReceiveEvent(tcpSetStateToEstablished, c.connMap)
			if err != nil {
				c.logger.Warn("error happened when receiving event:", zap.Error(err))
			}
			if stats != nil {
				ret = append(ret, stats)
			}
		} else if state == synSent || state == synRecv {
			continue
		} else {
			// These states are behind the ESTABLISHED state.
			// The codes could run into this branch if tcpSetStateToEstablished not received.
			c.logger.Debug("See sockets whose state is behind ESTABLISHED, which means no "+
				"tcp_set_state_from_established received.", zap.String("state", state))
			stats, err := connStat.StateMachine.ReceiveEvent(tcpSetStateFromEstablished, c.connMap)
			if err != nil {
				c.logger.Warn("error happened when receiving event:", zap.Error(err))
			}
			if stats != nil {
				ret = append(ret, stats)
			}
		}
	}
	return ret
}

func (c *ConnectMonitor) GetMapSize() int {
	return len(c.connMap)
}

func getConnKeyForTcpConnect(event *model.KindlingEvent) (ConnKey, error) {
	var sIpString string
	var sPortUint uint64
	var dIpString string
	var dPortUint uint64
	sIp := event.GetUserAttribute("sip")
	if sIp != nil {
		sIpString = model.IPLong2String(uint32(sIp.GetUintValue()))
	}
	sPort := event.GetUserAttribute("sport")
	if sPort != nil {
		sPortUint = sPort.GetUintValue()
	}
	dIp := event.GetUserAttribute("dip")
	if dIp != nil {
		dIpString = model.IPLong2String(uint32(dIp.GetUintValue()))
	}
	dPort := event.GetUserAttribute("dport")
	if dPort != nil {
		dPortUint = dPort.GetUintValue()
	}

	if sIp == nil || sPort == nil || dIp == nil || dPort == nil {
		return ConnKey{}, fmt.Errorf("some fields are nil for event %s. srcIp=%v, srcPort=%v, "+
			"dstIp=%v, dstPort=%v", event.Name, sIpString, sPortUint, dIpString, dPortUint)
	}

	connKey := ConnKey{
		SrcIP:   sIpString,
		SrcPort: uint32(sPortUint),
		DstIP:   dIpString,
		DstPort: uint32(dPortUint),
	}
	return connKey, nil
}
