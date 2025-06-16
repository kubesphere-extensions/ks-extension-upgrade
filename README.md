# KubeSphere Extension Upgrade

通用的 KubeSphere Extension 升级 Hook, 主要解决 **Extension 安装及升级时无法更新 CRD 问题**。当前支持以下特性：

- 支持扩展组件安装及升级时强制更新 CRD
- 扩展组件自定义支持
    - `whizard-monitoring` 从 1.1.x(1.0.x) 平滑升级至 1.2.x, 若启用旧版 Whizard 可观测中心，会将 whizard 剥离，并创建新扩展 `whizard-monitoring-pro`


### Quick start

> 版本要求： KSE >= v4.2.0, ks-extension-upgrade >= v0.3.0;

Extension 允许注入特定 Annotation， 在扩展安装及升级前执行自定义操作，其配置方式为镜像地址。 当您按照如下方式启用时, 在扩展部署阶段， `ks-extension-upgrade` 将作为 `helm-install(upgrade)-extension` Job 的 InitContainer 镜像，在扩展安装和升级之前, 执行强制更新 CRD 等操作，即扩展的 `pre-install(pre-upgrade) hook`.

#### 1. 修改配置，启用该扩展升级开关

修改 [defaultConfig](./pkg/config/config.go)，启用该扩展升级开关（参见 [PR](https://github.com/kubesphere-extensions/ks-extension-upgrade/pull/4) 实现）。另外也支持在扩展组件配置 *(values.yaml)* 中增加 `global.upgradeConfig.enabled` 配置项，来启用该特性。

```yaml
global:
  upgradeConfig:
    enabled: true
    installCrds: true
    upgradeCrds: true
    # mergeValues: false
    # failurePolicy: 0
    # dynamicOptions:
    #   key: value
```


#### 2. 为扩展组件增加特定 Annotations 

可在扩展的 extension.yaml 中注入如下 annotations，启用该特性。  

```
apiVersion: v2
name: whizard-monitoring
version: 1.2.0
displayName:
  zh: WizTelemetry 监控
  en: WizTelemetry Monitoring
annotations:
  executor-hook-image.kubesphere.io/install: kubesphere/ks-extension-upgrade:v0.3.0
  executor-hook-image.kubesphere.io/upgrade: kubesphere/ks-extension-upgrade:v0.3.0
```

### Issues

- 升级时配置合并是完全基于 [chartutil.MergeValues 函数](https://pkg.go.dev/helm.sh/helm/v3@v3.17.2/pkg/chartutil#MergeValues)， 实际部署时参数合并会更加复杂，请做好完备测试。