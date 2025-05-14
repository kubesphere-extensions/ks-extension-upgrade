# KubeSphere Extension Upgrade

通用的 KubeSphere Extension 升级 Hook, 主要解决 **Extension 安装及升级时无法更新 CRD 问题**。当前支持以下特性：

- 支持扩展组件安装及升级时强制更新 CRD
- 支持升级时将当前 InstallPlan Config 与升级版本的默认 Values 进行合并
- 扩展组件额外支持
    - `whizard-monitoring` 从 1.1.x 及 1.0.x 升级至 1.2.x 将 whizard 拆离，并创建新扩展 `whizard-monitoring-pro`


### Quick start

Extension 支持注入特定 Annotation， 在扩展安装及升级前执行自定义操作，在当前版本 (ks 4.1.x ~ 4.2) 配置为镜像地址形式。 当扩展在安装和升级之前， `ks-extension-upgrade` 将作为 `helm-install-extension`/`helm-upgrade-extension` Job 的 InitContainer 镜像 *(如 [Job 示例](./docs/helm-upgrade-extension-job.yaml) 所示)*， 执行强制更新 CRD。

#### 1. 修改配置，启用该扩展升级开关

修改 [defaultConfig](./pkg/config/config.go)，启用该扩展升级开关（参见 [PR](https://github.com/kubesphere-extensions/ks-extension-upgrade/pull/4) 实现）。另外可在**测试**或者**变更**场景中，在扩展部署或升级之前加载 [extension-upgrade-config.yaml](./docs/extension-upgrade-config.yaml) 配置文件，修改初始值。

#### 2. 在 Extension 配置中增加特定 Annotations 

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