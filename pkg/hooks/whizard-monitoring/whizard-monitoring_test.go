package whizardmonitoring

import (
	"testing"
)

func TestGenWhizardMonitoringSmoothUpgradeConfig(t *testing.T) {

	defaultconfig := `global:
  imageRegistry: ""
  imagePullSecrets: []
  clusterInfo: {}

whizard-monitoring-helper:
  etcdMonitoringHelper:
    enabled: false
  gpuMonitoringHelper:
    enabled: false

  hook:
    image:
      registry: docker.io
      repository: kubesphere/kubectl
      tag: v1.27.12

kubePrometheusStack:
  enabled: true

kube-prometheus-stack:

  prometheus:
    # agentMode need to be set to true when enable whizard
    agentMode: false

    prometheusSpec:
      image:
        registry: quay.io
        repository: prometheus/prometheus
        tag: v2.51.2
      replicas: 1
      retention: 7d
      resources:
        limits:
          cpu: "4"
          memory: 16Gi
        requests:
          cpu: 200m
          memory: 400Mi
      storageSpec:
        volumeClaimTemplate:
          spec:
            resources:
              requests:
                storage: 20Gi
      securityContext:
        fsGroup: 0
        runAsNonRoot: false
        runAsUser: 0
      nodeSelector: {}
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - preference:
              matchExpressions:
              - key: node-role.kubernetes.io/monitoring
                operator: Exists
            weight: 100
      tolerations: []
      secrets: []
      # - kube-etcd-client-certs ## be added when enable kubeEtcd servicemonitor with tls config

  prometheusOperator:
    image:
      registry: quay.io
      repository: prometheus-operator/prometheus-operator
      tag: v0.75.1

    
    admissionWebhooks:
      patch:
        image:
          registry: docker.io
          repository: kubespheredev/kube-webhook-certgen
          tag: v20221220-controller-v1.5.1-58-g787ea74b6

    prometheusConfigReloader:
      image:
        registry: quay.io
        repository: prometheus-operator/prometheus-config-reloader
        tag: v0.75.1
    
    nodeSelector: {}
    affinity: {}
    tolerations: []

  kube-state-metrics:
    image:
      registry: docker.io
      repository: kubesphere/kube-state-metrics
      tag: v2.12.0

    kubeRBACProxy:
      image:
        registry: quay.io
        repository: brancz/kube-rbac-proxy
        tag: v0.18.0

    nodeSelector:
      kubernetes.io/os: linux
    affinity: {}
    tolerations: []

  prometheus-node-exporter:
    image:
      registry: quay.io
      repository: prometheus/node-exporter
      tag: "v1.8.1"

    kubeRBACProxy:
      image:
        registry: quay.io
        repository: brancz/kube-rbac-proxy
        tag: v0.18.0

    ProcessExporter:
      enabled: false
      image:
        repository: kubesphere/process-exporter
        tag: "0.5.0"

    CalicoExporter:
      enabled: false
      image:
        repository: kubesphere/calico-exporter
        tag: v0.3.0

  kubeEtcd:
    ## If you want to enable etcd monitoring, set etcd endpoints and generate certificate secrets. The reference command is as follows:
    ##
    ## kubectl -n kubesphere-monitoring-system create secret generic kube-etcd-client-certs  \
    ## --from-file=etcd-client-ca.crt=/etc/ssl/etcd/ssl/ca.pem  \
    ## --from-file=etcd-client.crt=/etc/ssl/etcd/ssl/node-$(hostname).pem  \
    ## --from-file=etcd-client.key=/etc/ssl/etcd/ssl/node-$(hostname)-key.pem
    ##
    enabled: false
    endpoints: []
    #  - 172.31.73.206

dcgmExporter:
  enabled: false
  nodeSelector: {}

  image:
    registry: docker.io
    repository: kubesphere/dcgm-exporter
    tag: 3.3.5-3.4.0-ubuntu22.04
`

	t.Run("test gen whizard monitoring smooth upgrade config", func(t *testing.T) {
		config, err := genWhizardMonitoringSmoothUpgradeConfig(defaultconfig)
		if err != nil {
			t.Errorf("failed to gen whizard monitoring smooth upgrade config: %v", err)
		}
		t.Logf("config: %s", config)
	})
}
