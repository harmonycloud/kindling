# Copyright (C) 2022 Kindling Project Authors.
# SPDX-License-Identifier: Apache-2.0


/usr/bin/kindling-probe-loader

if [ -f "/opt/probe.o" ]; then
	export SYSDIG_BPF_PROBE="/opt/probe.o"
fi

/app/kindling-collector --config=/app/config/kindling-collector-config.yml