# KubeSphere Extension Upgrade

通用的 KubeSphere Extension 升级 Hook, 当前支持以下特性

- 支持安装及升级时强制加载 CRD
- 支持升级时将当前 InstallPlan Config 与升级版本的默认 Values 进行合并
- 支持通过配置文件进行参数控制
- 扩展组件额外支持
    - `whizard-monitoring` 从 1.1.x 及 1.0.x 升级至 1.2.x 将 whizard 拆离，并创建新扩展 `whizard-monitoring-pro`


### Quick start

#### 1. 修改默认配置，启用扩展升级

修改 [defaultConfig](./pkg/config/config.go)，默认启用该扩展升级开关，或者在后续通过加载的配置文件进行修改。

#### 2. 在 Extension 配置中增加特定 Annotations 

Extension 支持注入特定 Annotation，在当前版本(ks 4.1.x ~ 4.2)配置为镜像地址形式。 当扩展在安装和升级时， `ks-extension-upgrade` 将作为 `helm-install-extension`/`helm-upgrade-extension` Job 的 InitContainer， 在扩展部署/升级之前执行。

可在扩展的 extension.yaml 中注入如下 annotations，启用该特性。  

```
apiVersion: v2
name: whizard-monitoring
version: 1.1.0
displayName:
  zh: WizTelemetry 监控
  en: WizTelemetry Monitoring
annotations:
  executor-hook-image.kubesphere.io/install: kubesphere/ks-extension-upgrade:v0.1.0
  executor-hook-image.kubesphere.io/upgrade: kubesphere/ks-extension-upgrade:v0.1.0
```

### Issues

- 升级时配置合并是完全基于 [chartutil.MergeValues 函数](https://pkg.go.dev/helm.sh/helm/v3@v3.17.2/pkg/chartutil#MergeValues)， 实际部署时参数合并会更加复杂，请做好完备测试。