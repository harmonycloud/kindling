// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package k8sprocessor

import (
	"fmt"
	"os"
)

func getHostIpFromEnv() (string, error) {
	return getValueFromEnv("MY_NODE_IP", "unknown")
}

func getHostNameFromEnv() (string, error) {
	return getValueFromEnv("MY_NODE_NAME", "unknown")
}

func getValueFromEnv(env string, defaultValue string) (string, error) {
	value, ok := os.LookupEnv(env)
	if !ok {
		return defaultValue, fmt.Errorf("[%s] is not found in env variable which will be set [%s]", env, defaultValue)
	}
	return value, nil
}
