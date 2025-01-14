## 1.3.2 20220528

### en

- added and use v1beta3 as the crd storage version. changes since vebeta2 as follows:
  - TServer:
    - spec.k8s.readiness changed from string to array
    - added spec.k8s.command field
    - added spec.k8s.args field
  - TFrameworkConfig:
    - added ImageBuild.Executor.Image field
    - added ImageBuild.Executor.Secret field
    - changed ImageRegistry domain name to ImageUpload
- removed tdeploy crd and related code
- build the project with makefile
- fixed some tarscontroller server bugs
- added tarskaniko server for compiling images, and "Docker In Docker" is no longer required
- developed a new version and compatibility plan
- update tarsweb to v3.0.1

### cn

- 新增并将 v1beta3 作为 crd 存储版本, 相比 v1beta2 变动如下:
  - TServer:
    - spec.k8s.readiness 由 string 变更为 array
    - 新增 spec.k8s.command 域
    - 新增 spec.k8s.args 域
  - TFrameworkConfig:
    - 新增 ImageBuild.Executor.Image 字段
    - 新增 ImageBuild.Executor.Secret 字段
    - 变更 ImageRegistry 域名为 ImageUpload
- 移除了 tdeploy crd 及相关代码
- 使用 makefile 构建项目
- 修复了一些 tarscontroller 服务的 bug
- 新增 tarskaniko 服务用于编译镜像,且不再需要 "Docker In Docker"
- 制定了全新的版本与兼容性计划
- 更新 TarsWeb 到 v3.0.1

## v1.2.5 20220425

### en

- Update: tarscpp(v3.0.6) & tarsweb(v3.0.0)
- Add: Complete the support of cloud application market and add CRD: tframeworkkey yaml
- Add: replace ip in notify with domain
- Add: servantName can now be equal to "nodeobj"
- Add: add exec-build-cloud.sh to support cloud market

### cn

- Update: tarscpp(v3.0.6) & tarsweb(v3.0.0)
- 添加: 完成云应用市场的支持, 增加了 CRD: TFrameworkKey.yaml
- 添加: tarnotify 中替换 ip 为域名
- 添加: 业务服务的 obj 可以等于"nodeobj"
- 添加: 增加了 exec-build-cloud.sh 脚本以支持市场服务的打包

## v1.2.4 20220321

### en

- Update: tarscpp(v3.0.5) & tarsweb(v2.4.27)
- Add: ingress adapter all k8s version
- Fix: tarsnotify podName is empty

### cn

- Update: tarscpp(v3.0.5) & tarsweb(v2.4.27)
- 添加: ingress 适配了所有版本
- 修复: tarsnotify 修复 podName 是空的问题

## v1.2.3 20220305

### en

- TarsWeb update to v2.4.26
- tarsnotify add reportNotifyInfo

### cn

- TarsWeb 升级到 v2.4.26, 优化了 web 各种体验
- tarsnotify 添加了 reportNotifyInfo 接口

## v1.2.2 20220225

### en

- TarsWeb update to v2.4.25
- add unittest
- fix controller some bug

### cn

- TarsWeb 升级到 v2.4.25, 优化了 k8s 管理
- 增加了自动测试
- 修复 controller 的一些 bug

## v1.2.1 20220218

### en

- TarsWeb update to v2.4.24
- When tarsnode starts the CPP/go service, if no matching exe is found, look for the first server to start it
- The helm package has been upgraded and all configuration items supported by tserver have been added
- base-compiler-stretch is added to support lower versions of GCC
- fix controller some bug

### cn

- TarsWeb 升级到 v2.4.24, 优化了 k8s 管理
- tarsnode 启动 Cpp/Go 服务时, 如果找不到匹配的 exe, 则查找第一个 Server 来启动
- 升级了 helm 包, 增加了 TServer 所支持的所有配置项
- 增加了 base-compiler-stretch, 支持低版本的 gcc
- 修复 controller 的一些 bug

## v1.2.0 20220118

### en

- removed the v1alpha1 version of all crds
- upgrade crd, use v1beta2 as store version, change list:

  - TServer:
    - add tserver.k8s.imagePullPolicy
    - add tserver.k8s.updateStrategy
    - add tserver.k8s.launcherType
    - add tserver.k8s.release.nodeImage
    - add tserver.k8s.release.nodeSecret
    - add tars.io/MinReplicas annotation
    - add tars.io/MaxReplicas annotation
  - TEndpoint:
    - add tendpoint.release.nodeImage
    - add tendpoint.release.nodeSecret
    - add tendpoint.status.pods.pid
    - backport tendpoint.status.pods.pid to v1beta1
  - TDeploy
    - add tdeploy.apply.k8s.imagePullPolicy
    - add tdeploy.apply.k8s.updateStrategy
    - add tdeploy.apply.k8s.launcherType

- add TFrameworkConfig crd for framework config manage
- fix server start error in daemonset
- adjust timage.releases.id duplicate policy, same id then overwrite the old one
- tarsweb add pid show
- tarsweb add node image manage and framework config manage
- tarsweb use v2.4.23

### cn

- 所有 crd 移除了 v1alpha1 版本
- 所有 crd 新增并将 v1beta2 作为存储版本, 相比 v1beta1 变动如下:

  - TServer:
    - 新增 tserver.k8s.imagePullPolicy 字段
    - 新增 tserver.k8s.updateStrategy 字段
    - 新增 tserver.k8s.launcherType 字段
    - 新增 tserver.k8s.release.nodeImage 字段
    - 新增 tserver.k8s.release.nodeSecret 字段
    - 新增 tars.io/MinReplicas 注解
    - 新增 tars.io/MaxReplicas 注解
  - TEndpoint:
    - 新增 tendpoint.release.nodeImage 字段
    - 新增 tendpoint.release.nodeSecret 字段
    - 新增 tendpoint.status.pods.pid 字段
    - backport tendpoint.status.pods.pid 到 v1beta1
  - TDeploy
    - 新增 tdeploy.apply.k8s.imagePullPolicy 字段
    - 新增 tdeploy.apply.k8s.updateStrategy 字段
    - 新增 tdeploy.apply.k8s.launcherType 字段

- 新增 TFrameworkConfig crd,用于集群框架相关的配置
- 修复了 daemonset 模式下 启动失败的问题
- 调整了 timage.releases.id 重复的时的策略. 现在如果有相同的 id,会自动删除后者
- tarsweb 新增了 pid 显示
- tarsweb 新增了 node 镜像管理和 集群配置管理
- tarsweb 使用了 v2.4.23

## v1.1.1 20211205

### en

- The docker image is uniformly switched to the official Debian image
- The helm.build.id in tarsconroller & tarsframework yaml is set to latest by default
- The tarsimage process is optimized and logs are added
- Fixed secretname in tserver CRD
- Fixed the bug that tarsnode does not exit when the master control is not started
- Fixed php template
- Update charts of tarsconroller & tarsframework 的
- Tarscpp & tarsweb dependency changed to submodule mode
- Utf8 character set is used inside the image
- The time and time zone in the mirror are the same as that of the host
- es -> elk, Maintain consistency with tarsframework
- Fix tarsevent crash bug

### cn

- docker 镜像统一切换到官方的 debian 镜像
- tarsconroller & tarsframework yaml 中的 helm.build.id 缺省设置为 latest
- 优化了 tarsimage 的过程, 增加了日志
- 修复了 TServer crd 中的 secretName
- 修复了主控没启动时, tarsnode 不退出的 bug
- 修复了 php 的模板
- 更新了 tarsconroller & tarsframework 的 charts
- tarscpp & tarsweb 依赖改成了 submodule 模式
- 镜像内部都采用 utf8 字符集
- 镜像内时间和时区和宿主机相同
- 配置路径 es -> elk, 保持和 tarsframework 的一致性
- 修复 tarsevent 总是 crash 的 bug

## v1.1.0 20211024

### en

- tarsweb>=v2.4.19, tarscpp>=v2.4.22 or >= 3.0.1
- tarslog support call train
- CRD update. Note that when upgrading tars, you need to upgrade tarscontroller and CRD

### cn

- tarsweb>=v2.4.19, tarscpp>=v2.4.22 or >= 3.0.1
- tarslog 支持调用链
- crd 更新, 注意升级 Tars 时候需要升级 tarscontroller 以及更新 CRD

## v1.0.0 20210911

### en

- In the first version, tarsweb>=v2.4.18, tarscpp>=v2.4.22 or >= 3.0.0

### cn

- 第一个版本, tarsweb 要求>=v2.4.18, tarscpp 要求>=v2.4.22 or >= 3.0.0
