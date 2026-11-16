# Package Metaprovider

The Metaprovider package provides a component named `Metadata-provider` that runs externally. When the cluster is large, the Metadata-provider can replace the API-Server to provide K8s meta information to the kindling-agent to reduce the pressure on the API-Server.

The manifest file required for deployment is provided in the `deploy/metadata-provider/metadata-provider-deploy.yml`. Use the following command to start

```bash
kubectl create -f deploy/metadata-provider/metadata-provider-deploy.yml
```

The metadata-provider reads K8s metadata from the API-Server through the provided k8sAuth configuration (the kindling's ServiceAccount is used by default) and allows the agent to fetch K8s meta information

To make the agent use the service, you need to modify the configMap named `kindlingcfg`.

You can modify the configuration in `deploy/agent/kindling-collector-config.yml` before installation or directly modify the existing configMap named `kindlingcfg` .

```bash
kubectl edit configmap kindlingcfg -n kindling
```

The modified configuration is as follows:

- Modify `processors.k8smetadataprocessor.metadata_provider_config.enable` to true
- Modify `processors.k8smetadataprocessor.metadata_provider_config.endpoint` to the metadata-provider's service, which was created in the previous step

Examples are as follows:

```yaml
    processors:
        k8smetadataprocessor:
            # Set "enable" false if you want to run the agent in the non-Kubernetes environment.
            # Otherwise, the agent will panic if it can't connect to the API-server.
            enable: true
            #...
            metadata_provider_config:
                # confirm that metedata_provider is deployed before you enable the configuration
                # deploy metedata_provider by `kubectl create -f deploy/metadata-provider/metadata-provider-deploy.yml`
                enable: true
                # set `enable_trace` as true only if you need to debug the metadata from metadata_provider
                # each k8sMetadata fetched from metadata_provider will be printed into console
                enable_trace: false
                # check service endpoint by `kubectl get endpoints metadata-provider  -n kindling``
                endpoint: http://metadata-provider.kindling:9504
```

If you are modifying existing configmap resources, you need to restart all probes to apply the changes with command below:

```bash
kubectl delete po -l k8s-app=kindling-agent -n kindling
```
