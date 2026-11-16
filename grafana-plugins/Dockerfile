# Copyright (C) 2022-2023 Kindling Project Authors.
# SPDX-License-Identifier: Apache-2.0

FROM node:16.18.0-alpine3.15 AS pluginbuild
WORKDIR /app/topo-plugin
COPY topo-plugin .
RUN npm install
RUN npm run build

FROM grafana/grafana:8.5.26
COPY --from=pluginbuild /app/topo-plugin /var/lib/grafana/plugins/kindlingproject-topology-panel
COPY docker/grafana.ini /etc/grafana
COPY docker/dashboards.yml /etc/grafana/provisioning/dashboards
COPY dashboard-json /etc/grafana/dashboards-files
