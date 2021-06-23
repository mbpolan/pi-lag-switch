# pi-lag-switch

An experimental project which applies traffic shaping for your Raspberry Pi wifi access point.

# Prerequisites

This assumes you have a Raspberry Pi that has been configured as a 
[wireless access point](https://www.raspberrypi.org/documentation/configuration/wireless/access-point-bridged.md). This
was tested using a Raspberry Pi 4 Model B, set up using a bridged network configuration on the `wlan0` interface.

# Building

You will need at least Go 1.15. 

Static web assets are packed using the awesome [go.rice](https://github.com/GeertJohan/go.rice) package.
Install the package, and run the following command to generate a Go source file:

```bash
rice embed-go
```

Now build the project by running:

```bash
go build
```

Note: if you are building this on a macOS system, you can cross-compile for Linux ARM by using the following
command instead:

```bash
GOOS=linux GOARCH=arm go build
```

The build will produce a `lagswitch` binary, which you can run on your Raspberry PI (or equivalent Linux-based setup). 

# Running

Make sure your system already has the `tc` program available.

Since the `tc` program requires root privileges to change traffic rules, we'll need to allow the user that
runs the `lagswitch` process to do so as well. You _can_ use the default `pi` user gets created on most
Raspbian distro installations, which has `sudo` abilities out of the box, or you can create a user just
for this purpose.

```bash
$ sudo adduser lagswitch --system
```

Then, add the newly created system user to the sudoers file by running `sudo visudo`, and add the following
line to your file:

`lagswitch ALL=NOPASSWD: /sbin/tc`

To run the `lagswitch` program, you can provide an optional `-host` and `-port` argument to control what 
hostname or IP address and port number, respectively, the web interface will be available on. By default, 
the server will listen on all interfaces on port 8080.

```bash
$ ./lagswitch [-host] [-port]
```

All set! Open your browser to http://raspberrypi.local:8080 (or equivalent URL) and have fun.

# Cleaning Up

If you do not want to keep the traffic shaping rules applied by this project, run the following command
to remove all configuration changes that might be active:

```bash
$ sudo tc class delete dev wlan0 classid "1:10"
```

If you created a dedicated system user for this project, be sure to remove them as well.
