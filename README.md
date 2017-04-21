# LazyKube #

Easy deploy kuberentes

## Requirement ##

* docker
* openssh-client ( used to generate CA )
* qemu/KVM ( if deploy with qemu/KVM )
* libvirt ( if deploy with qemu/KVM )
* virst-install ( if deploy with qemu/KVM )

## How to deploy ##

Clone this repository and just using deploy and repository's home folder

```
git clone http://github.com/lyanchih/Lazykube
cd Lazykube
sudo ./scripts/deploy
```

The deploy script will check and execute few scripts to do deploy job.
You can study detail deploy steps at
[Detail deploy steps](scripts/README.md)

# lazy ini options #

Arrary's content is seperate by ","

## DEFAULT session ##

+--------------------+--------------------+--------------------+--------------------+--------------------+
|        key         |       value        |        type        |      require       |    description     |
+--------------------+--------------------+--------------------+--------------------+--------------------+
|       domain       |    example.com     |       string       |         *          |cluster node's base |
|                    |                    |                    |                    |       domain       |
+--------------------+--------------------+--------------------+--------------------+--------------------+
|      version       |      1235.9.0      |       string       |         *          |coreos image version|
+--------------------+--------------------+--------------------+--------------------+--------------------+
|      channel       |       stable       |       string       |         *          |coreos image channel|
|                    |                    |                    |                    |  (stable or dev)   |
+--------------------+--------------------+--------------------+--------------------+--------------------+
|        keys        |                    |      []string      |         *          | cluster node's ssh |
|                    |                    |                    |                    |     public key     |
+--------------------+--------------------+--------------------+--------------------+--------------------+
|       nodes        |                    |      []string      |         *          | cluster node list, |
|                    |                    |                    |                    |this will reference |
|                    |                    |                    |                    |  to node session   |
+--------------------+--------------------+--------------------+--------------------+--------------------+


## matchbox ##

+--------------------+------------------------+--------------------+--------------------+--------------------+
|        key         |       value            |        type        |      require       |    description     |
+--------------------+------------------------+--------------------+--------------------+--------------------+
|         ip         |     172.17.0.2         |       string       |         *          |matchbox container's|
|                    |                        |                    |                    |         IP         |
+--------------------+------------------------+--------------------+--------------------+--------------------+
|        url         |http://matchbox.com:8080|       string       |         *          |   matchbox's url   |
+--------------------+------------------------+--------------------+--------------------+--------------------+
|       domain       |      matchbox.com      |       string       |         *          | matchbox's domain  |
|                    |                        |                    |                    |   name, dns will   |
|                    |                        |                    |                    |record this address |
|                    |                        |                    |                    |       to IP        |
+--------------------+------------------------+--------------------+--------------------+--------------------+


## network ##

+--------------------+--------------------+--------------------+--------------------+-----------------------------+
|        key         |       value        |        type        |      require       |    description              |
+--------------------+--------------------+--------------------+--------------------+-----------------------------+
|      gateway       |     172.17.0.1     |       string       |         *          | Default router IP           |
+--------------------+--------------------+--------------------+--------------------+-----------------------------+
|        ips         |                    |      []string      |         *          |       Cluster network       |
|                    |                    |                    |                    |      settings, format:      |
|                    |                    |                    |                    |<cidr>:[<start_ip>[-<end_ip]]|
+--------------------+--------------------+--------------------+--------------------+-----------------------------+


## vip ##

+--------------------+--------------------+--------------------+--------------------+--------------------+
|        key         |       value        |        type        |      require       |    description     |
+--------------------+--------------------+--------------------+--------------------+--------------------+
|       enable       |        true        |      boolean       |         *          |     Enable vip     |
+--------------------+--------------------+--------------------+--------------------+--------------------+
|        vip         |    172.17.0.100    |       string       |         *          |        VIP         |
+--------------------+--------------------+--------------------+--------------------+--------------------+
|       domain       |  vip.cluster.com   |       string       |                    |     VIP domain     |
+--------------------+--------------------+--------------------+--------------------+--------------------+


## dns ##

+--------------------+--------------------+--------------------+--------------------+--------------------+
|        key         |       value        |        type        |      require       |    description     |
+--------------------+--------------------+--------------------+--------------------+--------------------+
|        dns         |      8.8.8.8       |      []string      |         *          | Cluster node's dns |
|                    |                    |                    |                    |      servers       |
+--------------------+--------------------+--------------------+--------------------+--------------------+


## nodes ##

+--------------------+--------------------+--------------------+--------------------+--------------------+
|        key         |       value        |        type        |      require       |    description     |
+--------------------+--------------------+--------------------+--------------------+--------------------+
|        mac         |                    |      []string      |         *          | Cluster node's mac |
|                    |                    |                    |                    |      address       |
+--------------------+--------------------+--------------------+--------------------+--------------------+
|        role        |                    |       string       |         *          |Cluster node's role |
+--------------------+--------------------+--------------------+--------------------+--------------------+


# LIMIT #

Currently only support deploy coreos
