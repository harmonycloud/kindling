// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package prometheusexporter

import (
	"fmt"
	"os"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"go.uber.org/zap"

	"go.opentelemetry.io/otel/attribute"
)

const (
	clusterIdEnv = "CLUSTER_ID"
	userIdEnv    = "USER_ID"

	CmonitorServiceNamePrefix = "cmonitor"
)

func getClusterIdFromEnv() (string, error) {
	return getValueFromEnv(clusterIdEnv, "noclusteridset")
}

func getUserIdFromEnv() (string, error) {
	return getValueFromEnv(userIdEnv, "nouseridset")
}

func getValueFromEnv(env string, defaultValue string) (string, error) {
	value, ok := os.LookupEnv(env)
	if !ok {
		return defaultValue, fmt.Errorf("[%s] is not found in env variable which will be set [%s]", env, defaultValue)
	}
	return value, nil
}

// GetHostname Return hostname if no error thrown, and set hostname 'unknown' otherwise.
func GetHostname() string {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "unknown"
	}
	return hostName
}

var commonLabels = []attribute.KeyValue{
	attribute.String("job", CmonitorServiceNamePrefix),
	attribute.String("instance", GetHostname()),
}

func GetCommonLabels(withUserInfo bool, logger *component.TelemetryLogger) []attribute.KeyValue {
	var clusterId, userId string
	var err error

	clusterId, err = getClusterIdFromEnv()
	if err != nil {
		logger.Error("Get ClusterId Failed", zap.Error(err))
	}
	userId, err = getUserIdFromEnv()
	if err != nil {
		logger.Error("Get UserId Failed", zap.Error(err))
	}
	if withUserInfo {
		return append(commonLabels,
			attribute.String("cluster_id", clusterId),
			attribute.String("user_id", userId))
	} else {
		return commonLabels
	}
}
