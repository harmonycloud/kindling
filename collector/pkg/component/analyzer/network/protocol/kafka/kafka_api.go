// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kafka

type apiVersion struct {
	minVersion int
	maxVersion int
}

const (
	_apiProduce                      = 0
	_apiFetch                        = 1
	_apiListOffsets                  = 2
	_apiMetadata                     = 3
	_apiLeaderAndIsr                 = 4
	_apiStopReplica                  = 5
	_apiUpdateMetadata               = 6
	_apiControlledShutdown           = 7
	_apiOffsetCommit                 = 8
	_apiOffsetFetch                  = 9
	_apiFindCoordinator              = 10
	_apiJoinGroup                    = 11
	_apiHeartbeat                    = 12
	_apiLeaveGroup                   = 13
	_apiSyncGroup                    = 14
	_apiDescribeGroups               = 15
	_apiListGroups                   = 16
	_apiSaslHandshake                = 17
	_apiApiVersions                  = 18
	_apiCreateTopics                 = 19
	_apiDeleteTopics                 = 20
	_apiDeleteRecords                = 21
	_apiInitProducerId               = 22
	_apiOffsetForLeaderEpoch         = 23
	_apiAddPartitionsToTxn           = 24
	_apiAddOffsetsToTxn              = 25
	_apiEndTxn                       = 26
	_apiWriteTxnMarkers              = 27
	_apiTxnOffsetCommit              = 28
	_apiDescribeAcls                 = 29
	_apiCreateAcls                   = 30
	_apiDeleteAcls                   = 31
	_apiDescribeConfigs              = 32
	_apiAlterConfigs                 = 33
	_apiAlterReplicaLogDirs          = 34
	_apiDescribeLogDirs              = 35
	_apiSaslAuthenticate             = 36
	_apiCreatePartitions             = 37
	_apiCreateDelegationToken        = 38
	_apiRenewDelegationToken         = 39
	_apiExpireDelegationToken        = 40
	_apiDescribeDelegationToken      = 41
	_apiDeleteGroups                 = 42
	_apiElectLeaders                 = 43
	_apiIncrementalAlterConfigs      = 44
	_apiAlterPartitionReassignments  = 45
	_apiListPartitionReassignments   = 46
	_apiOffsetDelete                 = 47
	_apiDescribeClientQuotas         = 48
	_apiAlterClientQuotas            = 49
	_apiDescribeUserScramCredentials = 50
	_apiAlterUserScramCredentials    = 51
	_apiAlterIsr                     = 56
	_apiUpdateFeatures               = 57
	_apiDescribeCluster              = 60
	_apiDescribeProducers            = 61
)

var kafka_apis = map[int]apiVersion{
	_apiProduce:                      {0, 9},
	_apiFetch:                        {0, 12},
	_apiListOffsets:                  {0, 7},
	_apiMetadata:                     {0, 11},
	_apiLeaderAndIsr:                 {0, 5},
	_apiStopReplica:                  {0, 3},
	_apiUpdateMetadata:               {0, 7},
	_apiControlledShutdown:           {0, 3},
	_apiOffsetCommit:                 {0, 8},
	_apiOffsetFetch:                  {0, 8},
	_apiFindCoordinator:              {0, 4},
	_apiJoinGroup:                    {0, 7},
	_apiHeartbeat:                    {0, 4},
	_apiLeaveGroup:                   {0, 4},
	_apiSyncGroup:                    {0, 5},
	_apiDescribeGroups:               {0, 5},
	_apiListGroups:                   {0, 4},
	_apiSaslHandshake:                {0, 1},
	_apiApiVersions:                  {0, 3},
	_apiCreateTopics:                 {0, 7},
	_apiDeleteTopics:                 {0, 6},
	_apiDeleteRecords:                {0, 2},
	_apiInitProducerId:               {0, 4},
	_apiOffsetForLeaderEpoch:         {0, 4},
	_apiAddPartitionsToTxn:           {0, 3},
	_apiAddOffsetsToTxn:              {0, 3},
	_apiEndTxn:                       {0, 3},
	_apiWriteTxnMarkers:              {0, 1},
	_apiTxnOffsetCommit:              {0, 3},
	_apiDescribeAcls:                 {0, 2},
	_apiCreateAcls:                   {0, 2},
	_apiDeleteAcls:                   {0, 2},
	_apiDescribeConfigs:              {0, 4},
	_apiAlterConfigs:                 {0, 2},
	_apiAlterReplicaLogDirs:          {0, 2},
	_apiDescribeLogDirs:              {0, 2},
	_apiSaslAuthenticate:             {0, 2},
	_apiCreatePartitions:             {0, 3},
	_apiCreateDelegationToken:        {0, 2},
	_apiRenewDelegationToken:         {0, 2},
	_apiExpireDelegationToken:        {0, 2},
	_apiDescribeDelegationToken:      {0, 2},
	_apiDeleteGroups:                 {0, 2},
	_apiElectLeaders:                 {0, 2},
	_apiIncrementalAlterConfigs:      {0, 1},
	_apiAlterPartitionReassignments:  {0, 0},
	_apiListPartitionReassignments:   {0, 0},
	_apiOffsetDelete:                 {0, 0},
	_apiDescribeClientQuotas:         {0, 1},
	_apiAlterClientQuotas:            {0, 1},
	_apiDescribeUserScramCredentials: {0, 0},
	_apiAlterUserScramCredentials:    {0, 0},
	_apiAlterIsr:                     {0, 0},
	_apiUpdateFeatures:               {0, 0},
	_apiDescribeCluster:              {0, 0},
	_apiDescribeProducers:            {0, 0},
}

func IsValidVersion(_api int, ver int) bool {
	version, ok := kafka_apis[_api]
	if !ok {
		return false
	}
	return version.minVersion <= ver && ver <= version.maxVersion
}
