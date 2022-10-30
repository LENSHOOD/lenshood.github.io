---
title: MacOS 上安装基于 QEMU 的 K8S 集群
date: 2022-09-20 09:44:23
tags:
- qemu
- macos
- k8s
categories:
- Kubernetes
---

{% asset_img header.jpg 500 %}

本文介绍了如何在 mac 上基于 qemu 来配置并启动一个 k8s 集群。

<!-- more -->

### QEMU 启动 VM

1. 安装 qemu

   `brew install qemu`

   qemu 安装完毕后，执行 `brew info qemu`，可以得知 qemu 安装在了 `/usr/local/Cellar/qemu/{version}` 目录下，进到 bin 目录下，我们会发现实际上安装了多个平台的 qemu，我所使用的是 MBP intel 版本，因此可以直接选用 `qemu-system-x86_64`， 能开启 mac 的硬件虚拟化。

2. 下载 Linux 映像

   这里使用的是 [CentOS Stream 8](https://www.centos.org/centos-stream/)，下载完成后得到文件：`CentOS-Stream-8-x86_64-latest-dvd1.iso`。

3. 为 vm 创建一个虚拟磁盘

   `qemu-img create $DISK_NAME $DISK_SIZE`

   DISK_NAME 是虚拟磁盘名称，这里DISK_NAME设置的是 `node0`，DISK_SIZE 是容量，这里选择的是 `15G` 足够 CentOS 的安装。

4. 挂载 iso 并启动 qemu

   ```shell
   qemu-system-x86_64 \
     -machine type=q35,accel=hvf \
     -smp 2 \
     -m 1G \
     -drive file=node0,index=0,media=disk \
     -cdrom CentOS-Stream-8-x86_64-latest-dvd1.iso
   ```

   关于 qemu 启动参数的详细见[这里](https://www.qemu.org/docs/master/system/invocation.html#hxtool-0)。

   简而言之，上述启动参数描述了一个 2C1G，虚拟磁盘文件名为 `node0`，启用了 mac 的 Hypervisor：hvf 加速器，且挂载了刚刚下载的 linux iso 映像的机器。

   一路安装完毕，就可以正常登录到 CentOS 中了。

5. 网络配置

   系统安装完毕后，我们会发现前面的配置中还少了一环，那就是网络。

   qemu 默认的 usermode 网络功能很有限，并且需要开启端口转发，我们更倾向于采用 bridge 的办法连通 guest 和 host 的网络。

   具体的方式就是通过在 host 上创建一个 bridge 网桥，与 qemu 创建的一个 TAP 虚拟设备连接起来，实现 qemu 程序与 host 网络的连接。

   具体步骤如下：

   1. 先创建一个 bridge，参考[苹果官方文档](https://support.apple.com/en-in/guide/mac-help/mh43557/mac)，之后在`System Preferences -> Sharing` 中，将`Internet Sharing` 配置和创建的网桥关联起来
   2. macOS 本身并不支持 tap/tun 虚拟设备，传统的做法是安装一个 [tuntaposx](http://tuntaposx.sourceforge.net/)：`brew install tuntap`。但由于 tuntaposx 已经归档不再更新，按照官网的指引，我们需要通过 [Tunnelblick]([https://tunnelblick.net](https://tunnelblick.net/)) 间接安装
   3. 在安装完成 Tunnelblick 之后，按照[其官方文档安装 tuntap](https://tunnelblick.net/cKextsInstallation.html)，过程中需要重启。
   4. 虽然 tuntap 安装完成了，但其内核扩展插件尚未加载，按照[这里的讨论](https://groups.google.com/g/tunnelblick-discuss/c/v5wnQCRZ8HY/m/Q8nhFBx1BgAJ)，我们通过如下命令来加载/卸载该内核扩展：
      - 加载：`/Applications/Tunnelblick.app/Contents/Resources/openvpnstart loadKexts 2`
      - 卸载：`/Applications/Tunnelblick.app/Contents/Resources/openvpnstart unloadKexts 2`

   5. 至此，我们可以为 qemu 添加网络相关的配置：

      ```shell
      qemu-system-x86_64 \
        -netdev tap,id=nd0,ifname=tap0,script=./qemu-ifup,downscript=./qemu-ifdown \
        -device e1000,netdev=nd0,mac=xx:xx:xx:xx:xx:xx
      ```

      其中 `-netdev` 定义了采用 tap 网络，并启动名为 tap0 的虚拟设备。`-device` 配置具体的设备，通过 `id=nd0` 与定义进行关联。如果不指定 mac 地址，则默认地址只有一个，如果要启动多个 vm 则会导致冲突。

      此外，`script` 和 `downscript` 分别配置两个脚本，在 qemu 启动、终止时执行。正好当 tap0 被 qemu 创建后，还没有和 bridge 做关联，所以 script 和 downscript 的内容可以分别为：`ifconfig bridge0 addm tap0` 和 `ifconfig bridge0 deletem tap0`（其中 `bridge0` 是我们在 mac 上创建的网桥名称）。

   6. 现在，重启 qemu，完整的启动命令：

      ```shell
      qemu-system-x86_64 \
        -machine type=q35,accel=hvf \
        -smp 2 \
        -m 1G \
        -drive file=node0,index=0,media=disk \
        -netdev tap,id=nd0,ifname=tap0,script=./qemu-ifup,downscript=./qemu-ifdown -device e1000,netdev=nd0
      ```

   7. 输入 `nmcli c reload` 重新加载网络连接，之后在 `ifconfig` 中就能看到网卡已经获取到了三层网络地址

   8. 为了方便下一次启动自动配置网络，在 `/etc/sysconfig/network-script/ifcfg-{nic_name}` 中配置 `ONBOOT = yes`



### 初始化 k8s 集群

1. 关闭前述 vm 的 swap：在 `/etc/fstab` 中将 swap 相关的行注释掉，之后重启。（关闭 swap 的主要原因是 swap 的存在让 kubelet 难以管理 pod 的内存使用，不过在 [v1.22 alpha 中已经尝试支持 swap](https://kubernetes.io/blog/2021/08/09/run-nodes-with-swap-alpha/)）

2. 安装 container runtime，这里选择的是安装 cri-o，[安装文档](https://github.com/cri-o/cri-o/blob/main/install.md#install-packaged-versions-of-cri-o)，完成后执行 `sudo systemctl start crio`

   > [覆写沙箱的 pause 镜像](registry.cn-hangzhou.aliyuncs.com/google_containers/pause)非常重要，如果忘记这一步，会导致 kubeadm 无法创建 control plane 的 pod。[pause (intra container)的作用？](https://jimmysong.io/kubernetes-handbook/concepts/pause-container.html)

3. 安装 kubeadm、kubelet、kubeadm

   可以选择国内 yum 源：

   ```shell
   cat <<EOF | sudo tee /etc/yum.repos.d/kubernetes.repo
   [kubernetes]
   name=Kubernetes
   baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
   enabled=1
   gpgcheck=1
   repo_gpgcheck=1
   gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
   EOF
   ```

   安装：`sudo yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes`

4. 可以正式开始启动 `kubeadm`

   ```shell
   kubeadm init \
     --pod-network-cidr=10.244.0.0/16 \
     --image-repository=registry.aliyuncs.com/google_containers \
     --kubernetes-version=stable \
     --cri-socket=unix:///var/run/crio/crio.sock
   ```

   可以看到这里使用了国内的镜像源，此外 cri-o 必须要配置其 socket。

   > 在这个过程中，遇到了 cri-o 在拉镜像时报的 `Unknown key "keyPaths"` 的问题，搜了下发现是一个[未修复的 bug](https://github.com/cri-o/cri-o/issues/6197#issuecomment-1236397274)，可以采用 issue 内的办法临时解决。

5. 现在按照 kubeadm 的提示，把 `/etc/kubernetes/admin.conf` 拷贝到 `$HOME/.kube/config` 之后就可以用 kubectl 连接集群了；另外，kubeadm 的输出中也包含了如何加入工作节点的命令，可以保留下来以备后面加入节点

6. 安装网络插件（这里选最简单的 Flannel）：

   ```shell
   kubectl apply \
     -f https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml
   ```

   当 kube-system ns 下出现 coredns 的 pod 时，说明网络插件安装成功了

7. 加入 Worker Node

   1. 以相同的方式初始化另一个 qemu vm

   2. 执行先前保存的加入命令，或是执行：

      `kubeadm join --token <token> <control-plane-host>:<control-plane-port> --discovery-token-ca-cert-hash sha256:<hash>` 

      命令内的参数如何获取见[这里](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/#join-nodes)

   3. 把 control plane 的 `/etc/kubernetes/admin.conf` 拷贝到 Worker Node，即可使用 kubectl



### Host 连接集群

创建 kubeconfig 来允许宿主机访问，流程参考[这里](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-certs/#kubeconfig-additional-users)

1. 生成当前集群的 kubeadm config 文件

   - 获取当前集群的 kubeadm 配置：

     `kubectl get cm kubeadm-config -n kube-system -o=jsonpath="{.data.ClusterConfiguration}"`

     并将其保存生成一个 kubeadm.conf

   - 使用 kubeadm 生成一个 kubeconfig 输出到 stdout：

     `kubeadm kubeconfig user --config kubeadm.conf --client-name outter-admin`  

     该命令生成了一个名为 `outter-admin` 的用户

   - 创建一个 clusterrolebinding，将 `outter-admin` 用户和内置的 clusterrole: `cluster-admin` 绑定起来

     `kubectl create clusterrolebinding outter-admin-for-cluster-admin --clusterrole=cluster-admin --user=outter-admin`

   - 更新宿主机上的 `~/.kube/config` 文件，将 kubeconfig 合并进去

2. 在宿主机验证

   - 执行 `kubectl get nodes` 就可以看到前文部署的两个节点
