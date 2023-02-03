# husarnet-dds

[![Firmware release](https://github.com/husarnet/husarnet-dds/actions/workflows/release.yaml/badge.svg)](https://github.com/husarnet/husarnet-dds/actions/workflows/release.yaml)

Automatically generate the DDS configuration for Husarnet.

## Installing from releases

```bash
RELEASE="v1.0.10"
ARCH="amd64"

sudo curl -L https://github.com/husarnet/husarnet-dds/releases/download/$RELEASE/husarnet-dds-linux-$ARCH -o /usr/local/bin/husarnet-dds
sudo chmod +x /usr/local/bin/husarnet-dds
```

## Building

```
go build -o husarnet-dds
```

## Using

> **warning**
>
> Target file name specified in `FASTRTPS_DEFAULT_PROFILES_FILE` or `CYCLONEDDS_URI` paths should contain 'husarnet' substring. Otherwise, the config will be saved under a default location.

### Fast DDS - simple discovery

```bash
export RMW_IMPLEMENTATION=rmw_fastrtps_cpp
export FASTRTPS_DEFAULT_PROFILES_FILE=./husarnet-fastdds-simple.xml
husarnet-dds singleshot
```

### Fast DDS - discovery server

#### Device acting as a **SERVER**

Let's assume that the Husarnet hostname of the device is `my-server`:

```bash
export DISCOVERY_SERVER_PORT=11811
husarnet-dds singleshot
```

Launching the discovery server:

```
fast-discovery-server -i 0 -x /var/tmp/husarnet-dds/fastdds-ds-server.xml
```

#### Device acting as a **CLINET**

```bash
export RMW_IMPLEMENTATION=rmw_fastrtps_cpp
export FASTRTPS_DEFAULT_PROFILES_FILE=./husarnet-fastdds-ds-client.xml
export ROS_DISCOVERY_SERVER=my-server:11811
husarnet-dds singleshot
```

### Cyclone DDS

```bash
export RMW_IMPLEMENTATION=rmw_cyclonedds_cpp
export CYCLONEDDS_URI=file://./husarnet-cyclone.xml
husarnet-dds singleshot
```


## Launching `husarnet-dds` as a service

You can run `husarnet-dds` in a service mode. It will start automatically after system reboot, and every 5 seconds it will update the hosts in DDS `.xml` profile files.

### Installing

```bash
export RMW_IMPLEMENTATION=rmw_fastrtps_cpp
export FASTRTPS_DEFAULT_PROFILES_FILE=/var/tmp/husarnet-dds/husarnet-fastdds.xml

sudo husarnet-dds install $USER \
-e RMW_IMPLEMENTATION=$RMW_IMPLEMENTATION \
-e FASTRTPS_DEFAULT_PROFILES_FILE=$FASTRTPS_DEFAULT_PROFILES_FILE
```

Available environtment variables

| key | default value | description |
| - | - | - |
| `RMW_IMPLEMENTATION` | `rmw_fastrtps_cpp` | Choosing a default DDS implementation. Possible values `rmw_fastrtps_cpp` and `rmw_cyclonedds_cpp` |
| `FASTRTPS_DEFAULT_PROFILES_FILE` | `/var/tmp/husarnet-dds/fastdds.xml` | path to the output `.xml` file if `RMW_IMPLEMENTATION=rmw_fastrtps_cpp` |
| `CYCLONEDDS_URI` | `/var/tmp/husarnet-dds/cyclonedds.xml` | path to the output `.xml` file if `RMW_IMPLEMENTATION=rmw_cyclonedds_cpp` |
| `ROS_DISCOVERY_SERVER` | (unset) | set it ONLY for devices running in [Fast DDS Discovery Server](https://fast-dds.docs.eprosima.com/en/latest/fastdds/ros2/discovery_server/ros2_discovery_server.html) **"Client mode"**. The value of this env should have the following format: `<Husarnet-hostname-of-discovery-server>:<PORT>`, eg. `my-ds-server:11811` |
| `DISCOVERY_SERVER_PORT` | `11811` | set it ONLY for devices running in [Fast DDS Discovery Server](https://fast-dds.docs.eprosima.com/en/latest/fastdds/ros2/discovery_server/ros2_discovery_server.html) **"Server mode"** |

### Starting

```bash
sudo husarnet-dds start
```

Now you can check the status of the service and last logs like that:

```bash
sudo systemctl status husarnet-dds.service 
```

```bash
sudo journalctl --unit husarnet-dds.service -n 100 
```

### Removing

```
sudo husarnet-dds stop
sudo husarnet-dds uninstall
```

## Creating a new release

```
git tag v1.0.0 main
git push origin v1.0.0 main
```