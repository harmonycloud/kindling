// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package k8sprocessor

import (
	"strconv"

	"go.uber.org/zap"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/processor"
	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"

	mpclient "github.com/Kindling-project/kindling/collector/pkg/metadata/metaprovider/client"
)

const (
	K8sMetadata = "k8smetadataprocessor"
	loopbackIp  = "127.0.0.1"
)

type K8sMetadataProcessor struct {
	config        *Config
	metadata      *kubernetes.K8sMetaDataCache
	nextConsumer  consumer.Consumer
	localNodeIp   string
	localNodeName string
	telemetry     *component.TelemetryTools
}

func NewKubernetesProcessor(cfg interface{}, telemetry *component.TelemetryTools, nextConsumer consumer.Consumer) processor.Processor {
	config, ok := cfg.(*Config)
	if !ok {
		telemetry.Logger.Panic("Cannot convert Component config", zap.String("componentType", K8sMetadata))
	}
	if !config.Enable {
		telemetry.Logger.Info("The kubernetes processor is disabled by the configuration. Won't connect to the API-server and no Kubernetes metadata will be fetched.")
		return &K8sMetadataProcessor{
			config:       config,
			metadata:     kubernetes.MetaDataCache,
			nextConsumer: nextConsumer,
			telemetry:    telemetry,
		}
	}
	var options []kubernetes.Option
	options = append(options,
		kubernetes.WithAuthType(config.KubeAuthType),
		kubernetes.WithKubeConfigDir(config.KubeConfigDir),
		kubernetes.WithGraceDeletePeriod(config.GraceDeletePeriod),
		kubernetes.WithFetchReplicaSet(config.EnableFetchReplicaSet),
	)
	if config.MetaDataProviderConfig != nil && config.MetaDataProviderConfig.Enable {
		cli := mpclient.NewMetaDataWrapperClient(config.MetaDataProviderConfig.Endpoint, config.MetaDataProviderConfig.EnableTrace)
		options = append(options, kubernetes.WithMetaDataProviderConfig(config.MetaDataProviderConfig, cli.ListAndWatch))
	}
	err := kubernetes.InitK8sHandler(options...)
	if err != nil {
		telemetry.Logger.Panicf("Failed to initialize [%s]: %v. Set the option 'enable' false if you want to run the agent in the non-Kubernetes environment.", K8sMetadata, err)
		return nil
	}

	var localNodeIp, localNodeName string
	if localNodeIp, err = getHostIpFromEnv(); err != nil {
		telemetry.Logger.Warn("Local NodeIp can not found", zap.Error(err))
	}
	if localNodeName, err = getHostNameFromEnv(); err != nil {
		telemetry.Logger.Warn("Local NodeName can not found", zap.Error(err))
	}
	return &K8sMetadataProcessor{
		config:        config,
		metadata:      kubernetes.MetaDataCache,
		nextConsumer:  nextConsumer,
		localNodeIp:   localNodeIp,
		localNodeName: localNodeName,
		telemetry:     telemetry,
	}
}

func (p *K8sMetadataProcessor) Consume(dataGroup *model.DataGroup) error {
	if !p.config.Enable {
		return p.nextConsumer.Consume(dataGroup)
	}
	name := dataGroup.Name
	switch name {
	case constnames.NetRequestMetricGroupName:
		p.processNetRequestMetric(dataGroup)
	case constnames.TcpRttMetricGroupName:
		fallthrough
	case constnames.TcpRetransmitMetricGroupName:
		fallthrough
	case constnames.TcpDropMetricGroupName:
		p.processTcpMetric(dataGroup)
	default:
		p.processNetRequestMetric(dataGroup)
	}
	return p.nextConsumer.Consume(dataGroup)
}

func (p *K8sMetadataProcessor) processNetRequestMetric(dataGroup *model.DataGroup) {
	isServer := dataGroup.Labels.GetBoolValue(constlabels.IsServer)
	if isServer {
		p.addK8sMetaDataForServerLabel(dataGroup.Labels)
	} else {
		p.addK8sMetaDataForClientLabel(dataGroup.Labels)
	}
}

func (p *K8sMetadataProcessor) processTcpMetric(dataGroup *model.DataGroup) {
	p.addK8sMetaDataViaIp(dataGroup.Labels)
}

func (p *K8sMetadataProcessor) addK8sMetaDataForClientLabel(labelMap *model.AttributeMap) {
	// add metadata for src
	containerId := labelMap.GetStringValue(constlabels.ContainerId)
	if containerId != "" {
		labelMap.UpdateAddStringValue(constlabels.SrcContainerId, containerId)
		resInfo, ok := p.metadata.GetByContainerId(containerId)
		if ok {
			addContainerMetaInfoLabelSRC(labelMap, resInfo)
		} else {
			labelMap.UpdateAddStringValue(constlabels.SrcNodeIp, p.localNodeIp)
			labelMap.UpdateAddStringValue(constlabels.SrcNode, p.localNodeName)
			labelMap.UpdateAddStringValue(constlabels.SrcNamespace, constlabels.InternalClusterNamespace)
		}
	} else {
		srcIp := labelMap.GetStringValue(constlabels.SrcIp)
		if srcIp == loopbackIp {
			labelMap.UpdateAddStringValue(constlabels.SrcNodeIp, p.localNodeIp)
			labelMap.UpdateAddStringValue(constlabels.SrcNode, p.localNodeName)
		}
		podInfo, ok := p.metadata.GetPodByIp(srcIp)
		if ok {
			addPodMetaInfoLabelSRC(labelMap, podInfo)
		} else {
			if nodeName, ok := p.metadata.GetNodeNameByIp(srcIp); ok {
				labelMap.UpdateAddStringValue(constlabels.SrcNodeIp, srcIp)
				labelMap.UpdateAddStringValue(constlabels.SrcNode, nodeName)
				labelMap.UpdateAddStringValue(constlabels.SrcNamespace, constlabels.InternalClusterNamespace)
			} else {
				labelMap.UpdateAddStringValue(constlabels.SrcNamespace, constlabels.ExternalClusterNamespace)
			}
		}
	}

	// add metadata for dst
	dstIp := labelMap.GetStringValue(constlabels.DstIp)
	if dstIp == loopbackIp {
		labelMap.UpdateAddStringValue(constlabels.DstNodeIp, p.localNodeIp)
		labelMap.UpdateAddStringValue(constlabels.DstNode, p.localNodeName)
		// If the dst IP is a loopback address, we use its src IP for further searching.
		dstIp = labelMap.GetStringValue(constlabels.SrcIp)
	}
	dstPort := labelMap.GetIntValue(constlabels.DstPort)
	// DstIp is IP of a service
	if svcInfo, ok := p.metadata.GetServiceByIpPort(dstIp, uint32(dstPort)); ok {
		labelMap.UpdateAddStringValue(constlabels.DstNamespace, svcInfo.Namespace)
		labelMap.UpdateAddStringValue(constlabels.DstService, svcInfo.ServiceName)
		labelMap.UpdateAddStringValue(constlabels.DstWorkloadKind, svcInfo.WorkloadKind)
		labelMap.UpdateAddStringValue(constlabels.DstWorkloadName, svcInfo.WorkloadName)
		// find podInfo using dnat_ip
		dNatIp := labelMap.GetStringValue(constlabels.DnatIp)
		dNatPort := labelMap.GetIntValue(constlabels.DnatPort)
		if dNatIp != "" && dNatPort != -1 {
			resInfo, ok := p.metadata.GetContainerByIpPort(dNatIp, uint32(dNatPort))
			if ok {
				addContainerMetaInfoLabelDST(labelMap, resInfo)
			} else {
				// maybe dnat_ip is NodeIP
				if nodeName, ok := p.metadata.GetNodeNameByIp(dNatIp); ok {
					labelMap.UpdateAddStringValue(constlabels.DstNodeIp, dNatIp)
					labelMap.UpdateAddStringValue(constlabels.DstNode, nodeName)
				}
			}
		}
	} else if resInfo, ok := p.metadata.GetContainerByIpPort(dstIp, uint32(dstPort)); ok {
		// DstIp is IP of a container
		addContainerMetaInfoLabelDST(labelMap, resInfo)
	} else if resInfo, ok := p.metadata.GetContainerByHostIpPort(dstIp, uint32(dstPort)); ok {
		addContainerMetaInfoLabelDST(labelMap, resInfo)
		labelMap.UpdateAddStringValue(constlabels.DstIp, resInfo.RefPodInfo.Ip)
		labelMap.UpdateAddIntValue(constlabels.DstPort, int64(resInfo.HostPortMap[int32(dstPort)]))
		labelMap.UpdateAddStringValue(constlabels.DstService, dstIp+":"+strconv.Itoa(int(dstPort)))
	} else {
		// DstIp is a IP from external
		if nodeName, ok := p.metadata.GetNodeNameByIp(dstIp); ok {
			labelMap.UpdateAddStringValue(constlabels.DstNodeIp, dstIp)
			labelMap.UpdateAddStringValue(constlabels.DstNode, nodeName)
			labelMap.UpdateAddStringValue(constlabels.DstNamespace, constlabels.InternalClusterNamespace)
		} else {
			labelMap.UpdateAddStringValue(constlabels.DstNamespace, constlabels.ExternalClusterNamespace)
		}
	}
}

func (p *K8sMetadataProcessor) addK8sMetaDataForServerLabel(labelMap *model.AttributeMap) {
	srcIp := labelMap.GetStringValue(constlabels.SrcIp)
	if srcIp == loopbackIp {
		labelMap.UpdateAddStringValue(constlabels.SrcNodeIp, p.localNodeIp)
		labelMap.UpdateAddStringValue(constlabels.SrcNode, p.localNodeName)
	}
	podInfo, ok := p.metadata.GetPodByIp(srcIp)
	if ok {
		addPodMetaInfoLabelSRC(labelMap, podInfo)
	} else {
		if nodeName, ok := p.metadata.GetNodeNameByIp(srcIp); ok {
			labelMap.UpdateAddStringValue(constlabels.SrcNodeIp, srcIp)
			labelMap.UpdateAddStringValue(constlabels.SrcNode, nodeName)
			labelMap.UpdateAddStringValue(constlabels.SrcNamespace, constlabels.InternalClusterNamespace)
		} else {
			labelMap.UpdateAddStringValue(constlabels.SrcNamespace, constlabels.ExternalClusterNamespace)
		}
	}
	containerId := labelMap.GetStringValue(constlabels.ContainerId)
	labelMap.UpdateAddStringValue(constlabels.DstContainerId, containerId)
	containerInfo, ok := p.metadata.GetByContainerId(containerId)
	if ok {
		addContainerMetaInfoLabelDST(labelMap, containerInfo)
		if containerInfo.RefPodInfo.ServiceInfo != nil {
			labelMap.UpdateAddStringValue(constlabels.DstService, containerInfo.RefPodInfo.ServiceInfo.ServiceName)
		}
	} else {
		labelMap.UpdateAddStringValue(constlabels.DstNodeIp, p.localNodeIp)
		labelMap.UpdateAddStringValue(constlabels.DstNode, p.localNodeName)
		labelMap.UpdateAddStringValue(constlabels.DstNamespace, constlabels.InternalClusterNamespace)
	}
}

// addK8sMetaDataViaIp is used to add k8s metadata to tcp metrics.
// There is also a piece of code for removing "port" in this method, which
// should be moved into a processor that is used for relabeling tcp metrics later.
func (p *K8sMetadataProcessor) addK8sMetaDataViaIp(labelMap *model.AttributeMap) {
	// Both Src and Dst should try:
	// 1. (Only Dst)Use Ip Port to find Service (when found a Service,also use DNatIp to find the Pod)
	// 2. Use Ip Port to find Container And Pod
	// 3. Use Ip to find Pod

	// add metadata for src
	p.addK8sMetaDataViaIpSRC(labelMap)
	// add metadata for dst
	p.addK8sMetaDataViaIpDST(labelMap)

	// We only care about the real connection, so here replace DstIp/DstPort with DNatIp/DNatPort
	if labelMap.GetStringValue(constlabels.DnatIp) != "" {
		labelMap.AddStringValue(constlabels.DstIp, labelMap.GetStringValue(constlabels.DnatIp))
	}
	if labelMap.GetIntValue(constlabels.DnatPort) > 0 {
		labelMap.AddIntValue(constlabels.DstPort, labelMap.GetIntValue(constlabels.DnatPort))
	}
	labelMap.RemoveAttribute(constlabels.DnatIp)
	labelMap.RemoveAttribute(constlabels.DnatPort)
	// Metric shouldn't contain high-cardinality labels, so here we want to remove
	// the dynamic port label and retain the listening one. But we can't know which
	// port is dynamic for sure, so we work around that by comparing their number size.
	//
	// The default dynamic port range in /proc/sys/net/ipv4/ip_local_port_range is 32768~60999.
	// At most cases, the larger port is the dynamic port and the other one is the listening port.
	// But sometimes the listening port is also greater than 32768 in which case there is no way to
	// tell which one is listening.
	var defaultMinLocalPort int64 = 32768
	srcPort := labelMap.GetIntValue(constlabels.SrcPort)
	dstPort := labelMap.GetIntValue(constlabels.DstPort)
	// If they are both smaller than 32768 then we remove the much smaller one.
	if srcPort < defaultMinLocalPort && dstPort < defaultMinLocalPort {
		if srcPort > dstPort {
			labelMap.RemoveAttribute(constlabels.SrcPort)
		} else {
			labelMap.RemoveAttribute(constlabels.DstPort)
		}
	} else {
		// Otherwise, we remove the port that is larger than 32768.
		if srcPort >= defaultMinLocalPort {
			labelMap.RemoveAttribute(constlabels.SrcPort)
		}
		if dstPort >= defaultMinLocalPort {
			labelMap.RemoveAttribute(constlabels.DstPort)
		}
	}
}

func (p *K8sMetadataProcessor) addK8sMetaDataViaIpSRC(labelMap *model.AttributeMap) {
	// 1. Use Ip Port to find Container And Pod
	// 2. Use Ip to find Pod
	srcIp := labelMap.GetStringValue(constlabels.SrcIp)
	srcPort := labelMap.GetIntValue(constlabels.SrcPort)
	srcContainerInfo, ok := p.metadata.GetContainerByIpPort(srcIp, uint32(srcPort))
	if ok {
		addContainerMetaInfoLabelSRC(labelMap, srcContainerInfo)
		return
	}

	srcPodInfo, ok := p.metadata.GetPodByIp(srcIp)
	if ok {
		addPodMetaInfoLabelSRC(labelMap, srcPodInfo)
		return
	}
	if _, ok := p.metadata.GetNodeNameByIp(srcIp); ok {
		labelMap.UpdateAddStringValue(constlabels.SrcNamespace, constlabels.InternalClusterNamespace)
	} else {
		labelMap.UpdateAddStringValue(constlabels.SrcNamespace, constlabels.ExternalClusterNamespace)
	}
}

func (p *K8sMetadataProcessor) addK8sMetaDataViaIpDST(labelMap *model.AttributeMap) {
	// 1. (Only Dst)Use Ip Port to find Service (when found a Service,also use DNatIp to find the Pod)
	// 2. Use Ip Port to find Container And Pod
	// 3. Use Ip to find Pod
	dstIp := labelMap.GetStringValue(constlabels.DstIp)
	dstPort := labelMap.GetIntValue(constlabels.DstPort)
	dstSvcInfo, ok := p.metadata.GetServiceByIpPort(dstIp, uint32(dstPort))
	if ok {
		labelMap.UpdateAddStringValue(constlabels.DstNamespace, dstSvcInfo.Namespace)
		labelMap.UpdateAddStringValue(constlabels.DstService, dstSvcInfo.ServiceName)
		labelMap.UpdateAddStringValue(constlabels.DstWorkloadKind, dstSvcInfo.WorkloadKind)
		labelMap.UpdateAddStringValue(constlabels.DstWorkloadName, dstSvcInfo.WorkloadName)
		// find podInfo using dnat_ip
		dNatIp := labelMap.GetStringValue(constlabels.DnatIp)
		dNatPort := labelMap.GetIntValue(constlabels.DnatPort)
		if dNatIp != "" && dNatPort != -1 {
			resInfo, ok := p.metadata.GetContainerByIpPort(dNatIp, uint32(dNatPort))
			if ok {
				addContainerMetaInfoLabelDST(labelMap, resInfo)
			}
		}
		return
	}

	dstContainerInfo, ok := p.metadata.GetContainerByIpPort(dstIp, uint32(dstPort))
	if ok {
		addContainerMetaInfoLabelDST(labelMap, dstContainerInfo)
		return
	}

	dstContainerInfo, ok = p.metadata.GetContainerByHostIpPort(dstIp, uint32(dstPort))
	if ok {
		addContainerMetaInfoLabelDST(labelMap, dstContainerInfo)
		labelMap.UpdateAddStringValue(constlabels.DstIp, dstContainerInfo.RefPodInfo.Ip)
		labelMap.UpdateAddIntValue(constlabels.DstPort, int64(dstContainerInfo.HostPortMap[int32(dstPort)]))
		labelMap.UpdateAddStringValue(constlabels.DstService, dstIp+":"+strconv.Itoa(int(dstPort)))
	}

	dstPodInfo, ok := p.metadata.GetPodByIp(dstIp)
	if ok {
		addPodMetaInfoLabelDST(labelMap, dstPodInfo)
		return
	}
	if _, ok := p.metadata.GetNodeNameByIp(dstIp); ok {
		labelMap.UpdateAddStringValue(constlabels.DstNamespace, constlabels.InternalClusterNamespace)
	} else {
		labelMap.UpdateAddStringValue(constlabels.DstNamespace, constlabels.ExternalClusterNamespace)
	}
}

func addContainerMetaInfoLabelSRC(labelMap *model.AttributeMap, containerInfo *kubernetes.K8sContainerInfo) {
	labelMap.UpdateAddStringValue(constlabels.SrcContainer, containerInfo.Name)
	labelMap.UpdateAddStringValue(constlabels.SrcContainerId, containerInfo.ContainerId)
	addPodMetaInfoLabelSRC(labelMap, containerInfo.RefPodInfo)
}

func addPodMetaInfoLabelSRC(labelMap *model.AttributeMap, podInfo *kubernetes.K8sPodInfo) {
	labelMap.UpdateAddStringValue(constlabels.SrcNode, podInfo.NodeName)
	labelMap.UpdateAddStringValue(constlabels.SrcNodeIp, podInfo.NodeAddress)
	labelMap.UpdateAddStringValue(constlabels.SrcNamespace, podInfo.Namespace)
	labelMap.UpdateAddStringValue(constlabels.SrcWorkloadKind, podInfo.WorkloadKind)
	labelMap.UpdateAddStringValue(constlabels.SrcWorkloadName, podInfo.WorkloadName)
	labelMap.UpdateAddStringValue(constlabels.SrcPod, podInfo.PodName)
	labelMap.UpdateAddStringValue(constlabels.SrcIp, podInfo.Ip)
	if podInfo.ServiceInfo != nil {
		labelMap.UpdateAddStringValue(constlabels.SrcService, podInfo.ServiceInfo.ServiceName)
	}
}

func addContainerMetaInfoLabelDST(labelMap *model.AttributeMap, containerInfo *kubernetes.K8sContainerInfo) {
	labelMap.UpdateAddStringValue(constlabels.DstContainer, containerInfo.Name)
	labelMap.UpdateAddStringValue(constlabels.DstContainerId, containerInfo.ContainerId)
	addPodMetaInfoLabelDST(labelMap, containerInfo.RefPodInfo)
}

func addPodMetaInfoLabelDST(labelMap *model.AttributeMap, podInfo *kubernetes.K8sPodInfo) {
	labelMap.UpdateAddStringValue(constlabels.DstNode, podInfo.NodeName)
	labelMap.UpdateAddStringValue(constlabels.DstNodeIp, podInfo.NodeAddress)
	labelMap.UpdateAddStringValue(constlabels.DstNamespace, podInfo.Namespace)
	labelMap.UpdateAddStringValue(constlabels.DstWorkloadKind, podInfo.WorkloadKind)
	labelMap.UpdateAddStringValue(constlabels.DstWorkloadName, podInfo.WorkloadName)
	labelMap.UpdateAddStringValue(constlabels.DstPod, podInfo.PodName)
	if labelMap.GetStringValue(constlabels.DstIp) == "" {
		labelMap.UpdateAddStringValue(constlabels.DstIp, podInfo.Ip)
	}
}
