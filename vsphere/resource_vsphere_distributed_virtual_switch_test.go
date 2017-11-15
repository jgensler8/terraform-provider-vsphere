package vsphere

import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/govmomi/vim25/types"
)

func TestAccResourceVSphereDistributedVirtualSwitch(t *testing.T) {
	var tp *testing.T
	testAccResourceVSphereDistributedVirtualSwitchCases := []struct {
		name     string
		testCase resource.TestCase
	}{
		{
			"basic",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfig(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
						),
					},
				},
			},
		},
		{
			"no hosts",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigNoHosts(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
						),
					},
				},
			},
		},
		{
			"remove a NIC",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfig(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
						),
					},
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigSingleNIC(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
						),
					},
				},
			},
		},
		{
			"standby with explicit failover order",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigStandbyLink(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchHasUplinks([]string{"tfup1", "tfup2"}),
							testAccResourceVSphereDistributedVirtualSwitchHasActiveUplinks([]string{"tfup1"}),
							testAccResourceVSphereDistributedVirtualSwitchHasStandbyUplinks([]string{"tfup2"}),
						),
					},
				},
			},
		},
		{
			"basic, then change to standby with failover order",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfig(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
						),
					},
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigStandbyLink(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchHasUplinks([]string{"tfup1", "tfup2"}),
							testAccResourceVSphereDistributedVirtualSwitchHasActiveUplinks([]string{"tfup1"}),
							testAccResourceVSphereDistributedVirtualSwitchHasStandbyUplinks([]string{"tfup2"}),
						),
					},
				},
			},
		},
		{
			"upgrade version",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigStaticVersion("6.0.0"),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchHasVersion("6.0.0"),
						),
					},
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigStaticVersion("6.5.0"),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchHasVersion("6.5.0"),
						),
					},
				},
			},
		},
		{
			"network resource control",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigNetworkResourceControl(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchHasNetworkResourceControlEnabled(),
							testAccResourceVSphereDistributedVirtualSwitchHasNetworkResourceControlVersion(
								types.DistributedVirtualSwitchNetworkResourceControlVersionVersion3,
							),
						),
					},
				},
			},
		},
		{
			"explicit uplinks",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigUplinks(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchHasUplinks([]string{"tfup1", "tfup2"}),
						),
					},
				},
			},
		},
		{
			"modify uplinks",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfig(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchHasUplinks(
								[]string{
									"uplink1",
									"uplink2",
									"uplink3",
									"uplink4",
								},
							),
						),
					},
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigStandbyLink(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchHasUplinks(
								[]string{
									"tfup1",
									"tfup2",
								},
							),
						),
					},
				},
			},
		},
		{
			"in folder",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigInFolder(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchMatchInventoryPath("tf-network-folder"),
						),
					},
				},
			},
		},
		{
			"single tag",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigSingleTag(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchCheckTags("terraform-test-tag"),
						),
					},
				},
			},
		},
		{
			"modify tags",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigSingleTag(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchCheckTags("terraform-test-tag"),
						),
					},
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigMultiTag(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchCheckTags("terraform-test-tags-alt"),
						),
					},
				},
			},
		},
		{
			"netflow",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigNetflow(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchHasNetflow(),
						),
					},
				},
			},
		},
		{
			"vlan ranges",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfigMultiVlanRange(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
							testAccResourceVSphereDistributedVirtualSwitchHasVlanRange(1000, 1999),
							testAccResourceVSphereDistributedVirtualSwitchHasVlanRange(3000, 3999),
						),
					},
				},
			},
		},
		{
			"import",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereDistributedVirtualSwitchPreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereDistributedVirtualSwitchConfig(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
						),
					},
					{
						ResourceName:      "vsphere_distributed_virtual_switch.dvs",
						ImportState:       true,
						ImportStateVerify: true,
						ImportStateIdFunc: func(s *terraform.State) (string, error) {
							dvs, err := testGetDVS(s, "dvs")
							if err != nil {
								return "", err
							}
							return dvs.InventoryPath, nil
						},
						Config: testAccResourceVSphereDistributedVirtualSwitchConfig(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereDistributedVirtualSwitchExists(true),
						),
					},
				},
			},
		},
	}

	for _, tc := range testAccResourceVSphereDistributedVirtualSwitchCases {
		t.Run(tc.name, func(t *testing.T) {
			tp = t
			resource.Test(t, tc.testCase)
		})
	}
}

func testAccResourceVSphereDistributedVirtualSwitchPreCheck(t *testing.T) {
	if os.Getenv("VSPHERE_HOST_NIC0") == "" {
		t.Skip("set VSPHERE_HOST_NIC0 to run vsphere_host_virtual_switch acceptance tests")
	}
	if os.Getenv("VSPHERE_HOST_NIC1") == "" {
		t.Skip("set VSPHERE_HOST_NIC1 to run vsphere_host_virtual_switch acceptance tests")
	}
	if os.Getenv("VSPHERE_ESXI_HOST") == "" {
		t.Skip("set VSPHERE_ESXI_HOST to run vsphere_host_virtual_switch acceptance tests")
	}
	if os.Getenv("VSPHERE_ESXI_HOST2") == "" {
		t.Skip("set VSPHERE_ESXI_HOST2 to run vsphere_host_virtual_switch acceptance tests")
	}
	if os.Getenv("VSPHERE_ESXI_HOST3") == "" {
		t.Skip("set VSPHERE_ESXI_HOST3 to run vsphere_host_virtual_switch acceptance tests")
	}
}

func testAccResourceVSphereDistributedVirtualSwitchExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dvs, err := testGetDVS(s, "dvs")
		if err != nil {
			if isAnyNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected DVS %s to be missing", dvs.Reference().Value)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasVersion(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		actual := props.Summary.ProductInfo.Version
		if expected != actual {
			return fmt.Errorf("expected version to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasNetworkResourceControlEnabled() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		actual := props.Config.(*types.VMwareDVSConfigInfo).NetworkResourceManagementEnabled
		if actual == nil || !*actual {
			return errors.New("expected network resource control to be enabled")
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasNetworkResourceControlVersion(expected types.DistributedVirtualSwitchNetworkResourceControlVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		actual := props.Config.(*types.VMwareDVSConfigInfo).NetworkResourceControlVersion
		if string(expected) != actual {
			return fmt.Errorf("expected network resource control version to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasUplinks(expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		policy := props.Config.(*types.VMwareDVSConfigInfo).UplinkPortPolicy.(*types.DVSNameArrayUplinkPortPolicy)
		actual := policy.UplinkPortName
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected uplinks to be %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasActiveUplinks(expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		pc := props.Config.(*types.VMwareDVSConfigInfo).DefaultPortConfig.(*types.VMwareDVSPortSetting)
		actual := pc.UplinkTeamingPolicy.UplinkPortOrder.ActiveUplinkPort
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected active uplinks to be %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasStandbyUplinks(expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		pc := props.Config.(*types.VMwareDVSConfigInfo).DefaultPortConfig.(*types.VMwareDVSPortSetting)
		actual := pc.UplinkTeamingPolicy.UplinkPortOrder.StandbyUplinkPort
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected standby uplinks to be %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasNetflow() resource.TestCheckFunc {
	expectedIPv4Addr := "10.0.0.100"
	expectedIpfixConfig := &types.VMwareIpfixConfig{
		CollectorIpAddress:  "10.0.0.10",
		CollectorPort:       9000,
		ObservationDomainId: 1000,
		ActiveFlowTimeout:   90,
		IdleFlowTimeout:     20,
		SamplingRate:        10,
		InternalFlowsOnly:   true,
	}

	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		actualIPv4Addr := props.Config.(*types.VMwareDVSConfigInfo).SwitchIpAddress
		actualIpfixConfig := props.Config.(*types.VMwareDVSConfigInfo).IpfixConfig

		if expectedIPv4Addr != actualIPv4Addr {
			return fmt.Errorf("expected switch IP to be %s, got %s", expectedIPv4Addr, actualIPv4Addr)
		}
		if !reflect.DeepEqual(expectedIpfixConfig, actualIpfixConfig) {
			return fmt.Errorf("expected netflow config to be %#v, got %#v", expectedIpfixConfig, actualIpfixConfig)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasVlanRange(emin, emax int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		pc := props.Config.(*types.VMwareDVSConfigInfo).DefaultPortConfig.(*types.VMwareDVSPortSetting)
		ranges := pc.Vlan.(*types.VmwareDistributedVirtualSwitchTrunkVlanSpec).VlanId
		var found bool
		for _, rng := range ranges {
			if rng.Start == emin && rng.End == emax {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("could not find start %d and end %d in %#v", emin, emax, ranges)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchMatchInventoryPath(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dvs, err := testGetDVS(s, "dvs")
		if err != nil {
			return err
		}

		expected, err := rootPathParticleNetwork.PathFromNewRoot(dvs.InventoryPath, rootPathParticleNetwork, expected)
		actual := path.Dir(dvs.InventoryPath)
		if err != nil {
			return fmt.Errorf("bad: %s", err)
		}
		if expected != actual {
			return fmt.Errorf("expected path to be %s, got %s", expected, actual)
		}
		return nil
	}
}

// testAccResourceVSphereDistributedVirtualSwitchCheckTags is a check to ensure that any tags
// that have been created with the supplied resource name have been attached to
// the folder.
func testAccResourceVSphereDistributedVirtualSwitchCheckTags(tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dvs, err := testGetDVS(s, "dvs")
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*VSphereClient).TagsClient()
		if err != nil {
			return err
		}
		return testObjectHasTags(s, tagsClient, dvs, tagResName)
	}
}

func testAccResourceVSphereDistributedVirtualSwitchConfig() string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

variable "esxi_hosts" {
  default = [
    "%s",
    "%s",
    "%s",
  ]
}

variable "network_interfaces" {
  default = [
    "%s",
    "%s",
  ]
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

data "vsphere_host" "host" {
  count         = "${length(var.esxi_hosts)}"
  name          = "${var.esxi_hosts[count.index]}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"

  host {
    host_system_id = "${data.vsphere_host.host.0.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.1.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.2.id}"
    devices = ["${var.network_interfaces}"]
  }
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
		os.Getenv("VSPHERE_ESXI_HOST"),
		os.Getenv("VSPHERE_ESXI_HOST2"),
		os.Getenv("VSPHERE_ESXI_HOST3"),
		os.Getenv("VSPHERE_HOST_NIC0"),
		os.Getenv("VSPHERE_HOST_NIC1"),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigStaticVersion(version string) string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

variable "esxi_hosts" {
  default = [
    "%s",
    "%s",
    "%s",
  ]
}

variable "network_interfaces" {
  default = [
    "%s",
    "%s",
  ]
}

variable "dvs_version" {
  default = "%s"
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

data "vsphere_host" "host" {
  count         = "${length(var.esxi_hosts)}"
  name          = "${var.esxi_hosts[count.index]}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
  version       = "${var.dvs_version}"

  host {
    host_system_id = "${data.vsphere_host.host.0.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.1.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.2.id}"
    devices = ["${var.network_interfaces}"]
  }
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
		os.Getenv("VSPHERE_ESXI_HOST"),
		os.Getenv("VSPHERE_ESXI_HOST2"),
		os.Getenv("VSPHERE_ESXI_HOST3"),
		os.Getenv("VSPHERE_HOST_NIC0"),
		os.Getenv("VSPHERE_HOST_NIC1"),
		version,
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigSingleNIC() string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

variable "esxi_hosts" {
  default = [
    "%s",
    "%s",
    "%s",
  ]
}

variable "network_interfaces" {
  default = [
    "%s",
  ]
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

data "vsphere_host" "host" {
  count         = "${length(var.esxi_hosts)}"
  name          = "${var.esxi_hosts[count.index]}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"

  host {
    host_system_id = "${data.vsphere_host.host.0.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.1.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.2.id}"
    devices = ["${var.network_interfaces}"]
  }
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
		os.Getenv("VSPHERE_ESXI_HOST"),
		os.Getenv("VSPHERE_ESXI_HOST2"),
		os.Getenv("VSPHERE_ESXI_HOST3"),
		os.Getenv("VSPHERE_HOST_NIC0"),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigNetworkResourceControl() string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

variable "esxi_hosts" {
  default = [
    "%s",
    "%s",
    "%s",
  ]
}

variable "network_interfaces" {
  default = [
    "%s",
  ]
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

data "vsphere_host" "host" {
  count         = "${length(var.esxi_hosts)}"
  name          = "${var.esxi_hosts[count.index]}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"

  network_resource_control_enabled = true
  network_resource_control_version = "version3"

  host {
    host_system_id = "${data.vsphere_host.host.0.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.1.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.2.id}"
    devices = ["${var.network_interfaces}"]
  }
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
		os.Getenv("VSPHERE_ESXI_HOST"),
		os.Getenv("VSPHERE_ESXI_HOST2"),
		os.Getenv("VSPHERE_ESXI_HOST3"),
		os.Getenv("VSPHERE_HOST_NIC0"),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigUplinks() string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

variable "esxi_hosts" {
  default = [
    "%s",
    "%s",
    "%s",
  ]
}

variable "network_interfaces" {
  default = [
    "%s",
  ]
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

data "vsphere_host" "host" {
  count         = "${length(var.esxi_hosts)}"
  name          = "${var.esxi_hosts[count.index]}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"

  uplinks = ["tfup1", "tfup2"]

  host {
    host_system_id = "${data.vsphere_host.host.0.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.1.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.2.id}"
    devices = ["${var.network_interfaces}"]
  }
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
		os.Getenv("VSPHERE_ESXI_HOST"),
		os.Getenv("VSPHERE_ESXI_HOST2"),
		os.Getenv("VSPHERE_ESXI_HOST3"),
		os.Getenv("VSPHERE_HOST_NIC0"),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigStandbyLink() string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

variable "esxi_hosts" {
  default = [
    "%s",
    "%s",
    "%s",
  ]
}

variable "network_interfaces" {
  default = [
    "%s",
  ]
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

data "vsphere_host" "host" {
  count         = "${length(var.esxi_hosts)}"
  name          = "${var.esxi_hosts[count.index]}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"

  uplinks         = ["tfup1", "tfup2"]
  active_uplinks  = ["tfup1"]
  standby_uplinks = ["tfup2"]

  host {
    host_system_id = "${data.vsphere_host.host.0.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.1.id}"
    devices = ["${var.network_interfaces}"]
  }

  host {
    host_system_id = "${data.vsphere_host.host.2.id}"
    devices = ["${var.network_interfaces}"]
  }
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
		os.Getenv("VSPHERE_ESXI_HOST"),
		os.Getenv("VSPHERE_ESXI_HOST2"),
		os.Getenv("VSPHERE_ESXI_HOST3"),
		os.Getenv("VSPHERE_HOST_NIC0"),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigNoHosts() string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigInFolder() string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

resource "vsphere_folder" "folder" {
  path          = "tf-network-folder"
  type          = "network"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
  folder        = "${vsphere_folder.folder.path}"
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigSingleTag() string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

resource "vsphere_tag_category" "terraform-test-category" {
  name        = "terraform-test-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "VmwareDistributedVirtualSwitch",
  ]
}

resource "vsphere_tag" "terraform-test-tag" {
  name        = "terraform-test-tag"
  category_id = "${vsphere_tag_category.terraform-test-category.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
  tags          = ["${vsphere_tag.terraform-test-tag.id}"]
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigMultiTag() string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

variable "extra_tags" {
  default = [
    "terraform-test-thing1",
    "terraform-test-thing2",
  ]
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

resource "vsphere_tag_category" "terraform-test-category" {
  name        = "terraform-test-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "VmwareDistributedVirtualSwitch",
  ]
}

resource "vsphere_tag" "terraform-test-tag" {
  name        = "terraform-test-tag"
  category_id = "${vsphere_tag_category.terraform-test-category.id}"
}

resource "vsphere_tag" "terraform-test-tags-alt" {
  count       = "${length(var.extra_tags)}"
  name        = "${var.extra_tags[count.index]}"
  category_id = "${vsphere_tag_category.terraform-test-category.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
  tags          = ["${vsphere_tag.terraform-test-tags-alt.*.id}"]
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigNetflow() string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"

  ipv4_address                  = "10.0.0.100"
  netflow_enabled               = true
  netflow_active_flow_timeout   = 90
  netflow_collector_ip_address  = "10.0.0.10"
  netflow_collector_port        = 9000
  netflow_idle_flow_timeout     = 20
  netflow_internal_flows_only   = true
  netflow_observation_domain_id = 1000
  netflow_sampling_rate         = 10
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigMultiVlanRange() string {
	return fmt.Sprintf(`
variable "datacenter" {
  default = "%s"
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacenter}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "terraform-test-dvs"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"

  vlan_range {
    min_vlan = 1000
    max_vlan = 1999
  }

  vlan_range {
    min_vlan = 3000
    max_vlan = 3999
  }
}
`,
		os.Getenv("VSPHERE_DATACENTER"),
	)
}
