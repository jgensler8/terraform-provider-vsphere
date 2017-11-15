---
layout: "vsphere"
page_title: "VMware vSphere: vsphere_network"
sidebar_current: "docs-vsphere-data-source-network"
description: |-
  Provides a vSphere network data source. This can be used to get the general attributes of a vSphere network.
---

# vsphere\_network

The `vsphere_network` data source can be used to discover the ID of a network
in vSphere. This can be any network that can be used as the backing for a
network interface for `vsphere_virtual_machine` or any other vSphere resource
that requires a network. This includes standard (host-based) port groups, DVS
port groups, or opaque networks such as those managed by NSX.

~> **NOTE:** This data source requires vCenter and is not available on direct
ESXi connections.

## Example Usage

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc1"
}

data "vsphere_network" "net" {
  name          = "terraform-test-net"
  datacenter_id = "${data.vsphere_datacenter.datacenter.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the network. This can be a name or path.
* `datacenter_id` - (Optional) The managed object reference ID of the
  datacenter the network is located in. This can be omitted if the search path
  used in `name` is an absolute path, or if there is only one datacenter in the
  vSphere infrastructure.

## Attribute Reference

The following attributes are exported:

* `id`: The managed object ID of the network in question.
* `type`: The managed object type for the discovered network. This will be one
  of `DistributedVirtualPortgroup` for DVS port groups, `Network` for standard
  (host-based) port groups, or `OpaqueNetwork` for networks managed externally
  by features such as NSX.
