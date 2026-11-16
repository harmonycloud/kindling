## This Topology Plugin is desgin for project Kindling?
This is a topology component based on ANTV developed specifically for the Kindling project. Prometheus data is queried using modified EBPF probes, so this topology is not a generic plugin.
To use this component together with [Kindling](https://github.com/KindlingProject/kindling), you need to import a kindling customized dashboard([dashboard.json](https://github.com/KindlingProject/kindling/blob/main/grafana-plugins/dashboard-json/topology.json)) that goes back to Prometheus to query the data collected by the probe.


After integrating this Grafana plugin and adding the corresponding Prometheus data source, you will see the plugin as shown
![img](https://raw.githubusercontent.com/thousandxu/zipImage/main/topo-plugin/topo.png) in `Kindling
Topology` dashboard

## What does this plugin do？
Topology call relationships are generated based on data queried in Prometheus and can be aggregated based on namespace and workload. The topology call relationship displays indicators such as Latency、calls、error rate and Volume. You can also view indicator data of nodes。

## How to build

Clone the [kindling project](https://github.com/KindlingProject/kindling) and navigate to this directory.

### Build topo plugin with docker

```shell
docker run -it -v $PWD:/topo-plugin node:16-bullseye /bin/bash
# Run the following command in the container's bash shell
cd /topo-plugin
yarn install
yarn build
# Then exit the container
exit
```

Check you get following yields

```
dist/
yarn.lock
coverage/
node_modules/
```

### Integrate this plugin with Grafana

```shell
# remove trashy yields
rm -rf yarn.lock coverage/ node_modules/
# build your own grafana-with-topo-plugin image
cd .. && docker build . -t grafana-with-topo-plugin
```
