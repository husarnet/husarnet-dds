# husarnet-dds

[![Firmware release](https://github.com/husarnet/husarnet-dds/actions/workflows/release.yaml/badge.svg)](https://github.com/husarnet/husarnet-dds/actions/workflows/release.yaml)

Automatically generate the DDS configuration for Husarnet.

## Installing from releases

### Linux/MacOS

Download the binary using `curl` or `wget`, which are available on most systems either preinstalled, or obtainable via package manager:

```bash
RELEASE="v1.3.5"
ARCH="amd64"

sudo curl -L https://github.com/husarnet/husarnet-dds/releases/download/$RELEASE/husarnet-dds-linux-$ARCH -o /usr/local/bin/husarnet-dds
sudo chmod +x /usr/local/bin/husarnet-dds
```

### Windows

On Windows, you can download the binary via GUI from [Releases page](https://github.com/husarnet/husarnet-dds/releases) and place in directory of your liking.
To do it right from PowerShell, use `wget` command which is an alias for `Invoke-WebRequest` cmdlet:

```
wget https://github.com/husarnet/husarnet-dds/releases/download/v1.3.4/husarnet-dds-windows-amd64.exe -OutFile husarnet-dds.exe
```

In order to be able to run the binary from any directory, place it in the directory present in your %PATH% environment variable.

## Building

```
go build -o husarnet-dds
```

## Using

> **warning**
>
> Target file name specified in `FASTRTPS_DEFAULT_PROFILES_FILE` or `CYCLONEDDS_URI` paths should contain 'husarnet' substring. Otherwise, the config will be saved under a default location.

### Fast DDS - simple discovery

On Linux/MacOS, using bash:

```bash
export RMW_IMPLEMENTATION=rmw_fastrtps_cpp
export FASTRTPS_DEFAULT_PROFILES_FILE=./husarnet-fastdds-simple.xml
husarnet-dds singleshot
```

On Windows, using PowerShell

```powershell
$env:RMW_IMPLEMENTATION = 'rmw_fastrtps_cpp'
$env:FASTRTPS_DEFAULT_PROFILES_FILE = '.\husarnet-fastdds-simple.xml'
.\husarnet-dds.exe singleshot
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

#### Device acting as a **CLIENT**

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

Unix-like systems, using bash:

```bash
export RMW_IMPLEMENTATION=rmw_fastrtps_cpp
export FASTRTPS_DEFAULT_PROFILES_FILE=/var/tmp/husarnet-dds/husarnet-fastdds.xml

sudo husarnet-dds install $USER \
-e RMW_IMPLEMENTATION=$RMW_IMPLEMENTATION \
-e FASTRTPS_DEFAULT_PROFILES_FILE=$FASTRTPS_DEFAULT_PROFILES_FILE
```

On Windows, using PowerShell as Administrator:

```powershell
$env:RMW_IMPLEMENTATION = 'rmw_fastrtps_cpp'
$env:FASTRTPS_DEFAULT_PROFILES_FILE = 'C:\Users\husarnetman\Desktop\dds\husarnet-fastdds-simple.xml'

.\husarnet-dds.exe install LocalSystem -e RMW_IMPLEMENTATION=$env:RMW_IMPLEMENTATION `
-e FASTRTPS_DEFAULT_PROFILES_FILE=$env:FASTRTPS_DEFAULT_PROFILES_FILE
```

Be sure to provide absolute path to the file which will be watched, as in the example.

Available environment variables

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

Now you can check the status of the service on Linux and last logs like that:

```bash
sudo systemctl status husarnet-dds.service 
```

```bash
sudo journalctl --unit husarnet-dds.service -n 100 
```

On Windows, see services.msc

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
