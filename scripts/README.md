# LazyKube
Deploy kuberentes detail steps

## Create containers

LazyKube deploy tools needs two service
One is matchbox, and the other is dnsmasq

matchbox used for pxe and cloud-init information
dnsmasq used for dhcp, tftp and dns service

We can deploy these two service by docker with following command

```
./scripts/docker-deploy
```

Then you will got one IP address, which is the matchbox service address
Please don't forget that, you will need that when you config your deploy config

## Generate cluster config

### configure your lazy ini file

You need to configure your cluster config.
This project had offer default ini file which is located at etc/lazy.ini

> Don't forget to place your ssh key at [DEFAULT]/keys

### build lazykube binary file

There are two methods to do this

If you had installed golang, you can just make binary file which will default
stored at _bin folder

```
make build
```

Or using container to make binary file

```
make container_build
```

### generate cluster config

Just run lazykube execute file, output files will default stored at _output

```
./_bin/lazykube config
```

Or you can see usage

```
./bin/lazykube help
```

### restart dnsmasq service

If you had change the config, don't forget to restart dnsmasq service
Or the new dns informaion will not work.
You can just retype previous command

```
./scrips/docker-deploy
```

### download coreos image

Currently lazykube only support coreos, but matchbox only using ipxe or pxe
to boot physical machine. If you have other os kernel and initramfs, of course
you canusing that. However the simplest way is just download coreos.

```
./scripts/get-coreos stable 1235.9.0
```

### generate tls certificate

There is one script can generate tls cetificate and it will parse your cluster
config. Just do not forget to regenerate certificate before you begin to deploy

```
./scripts/tls_gen.sh
```

### boot your machine

The most simple thing is using libvirt, we can just using following command
to create VMs. Script will automatic parse lazy ini file to operate VMs.

```
./scripts/libvirt create
```

And destroy VMs

```
./scripts/libvirt destroy
```

You can get usage about the script by

```
./scripts/libvirt -h
```
