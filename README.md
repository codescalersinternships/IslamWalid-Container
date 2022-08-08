# Container
Simple Container implementation in go using linux namespaces and cgroups.

## Usage:
```
git clone https://github.com/codescalersinternships/IslamWalid-container.git
cd IslamWalid-container
mkdir rootfs
mkdir rootfs/dev rootfs/bin rootfs/proc
./busybox --install ./rootfs/bin
sudo ./container run <command>
```
