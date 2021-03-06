package vsphere

import (
	"fmt"

	"context"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVSphereHostVirtualSwitch() *schema.Resource {
	s := map[string]*schema.Schema{
		"name": &schema.Schema{
			Type:        schema.TypeString,
			Description: "The name of the virtual switch.",
			Required:    true,
			ForceNew:    true,
		},
		"host_system_id": &schema.Schema{
			Type:        schema.TypeString,
			Description: "The managed object ID of the host to set the virtual switch up on.",
			Required:    true,
			ForceNew:    true,
		},
	}
	mergeSchema(s, schemaHostVirtualSwitchSpec())

	// Transform any necessary fields in the schema that need to be updated
	// specifically for this resource.
	s["active_nics"].Required = true
	s["standby_nics"].Required = true

	s["teaming_policy"].Default = hostNetworkPolicyNicTeamingPolicyModeLoadbalanceSrcID
	s["check_beacon"].Default = false
	s["notify_switches"].Default = true
	s["failback"].Default = true

	s["allow_promiscuous"].Default = false
	s["allow_forged_transmits"].Default = true
	s["allow_mac_changes"].Default = true

	s["shaping_enabled"].Default = false

	return &schema.Resource{
		Create: resourceVSphereHostVirtualSwitchCreate,
		Read:   resourceVSphereHostVirtualSwitchRead,
		Update: resourceVSphereHostVirtualSwitchUpdate,
		Delete: resourceVSphereHostVirtualSwitchDelete,
		Schema: s,
	}
}

func resourceVSphereHostVirtualSwitchCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*VSphereClient).vimClient
	name := d.Get("name").(string)
	hsID := d.Get("host_system_id").(string)
	ns, err := hostNetworkSystemFromHostSystemID(client, hsID)
	if err != nil {
		return fmt.Errorf("error loading host network system: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	spec := expandHostVirtualSwitchSpec(d)
	if err := ns.AddVirtualSwitch(ctx, name, spec); err != nil {
		return fmt.Errorf("error adding host vSwitch: %s", err)
	}

	saveHostVirtualSwitchID(d, hsID, name)

	return resourceVSphereHostVirtualSwitchRead(d, meta)
}

func resourceVSphereHostVirtualSwitchRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*VSphereClient).vimClient
	hsID, name, err := virtualSwitchIDsFromResourceID(d)
	if err != nil {
		return err
	}
	ns, err := hostNetworkSystemFromHostSystemID(client, hsID)
	if err != nil {
		return fmt.Errorf("error loading host network system: %s", err)
	}

	sw, err := hostVSwitchFromName(client, ns, name)
	if err != nil {
		return fmt.Errorf("error fetching virtual switch data: %s", err)
	}

	if err := flattenHostVirtualSwitchSpec(d, &sw.Spec); err != nil {
		return fmt.Errorf("error setting resource data: %s", err)
	}

	return nil
}

func resourceVSphereHostVirtualSwitchUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*VSphereClient).vimClient
	hsID, name, err := virtualSwitchIDsFromResourceID(d)
	if err != nil {
		return err
	}
	ns, err := hostNetworkSystemFromHostSystemID(client, hsID)
	if err != nil {
		return fmt.Errorf("error loading host network system: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	spec := expandHostVirtualSwitchSpec(d)
	if err := ns.UpdateVirtualSwitch(ctx, name, *spec); err != nil {
		return fmt.Errorf("error updating host vSwitch: %s", err)
	}

	return resourceVSphereHostVirtualSwitchRead(d, meta)
}

func resourceVSphereHostVirtualSwitchDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*VSphereClient).vimClient
	hsID, name, err := virtualSwitchIDsFromResourceID(d)
	if err != nil {
		return err
	}
	ns, err := hostNetworkSystemFromHostSystemID(client, hsID)
	if err != nil {
		return fmt.Errorf("error loading host network system: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	if err := ns.RemoveVirtualSwitch(ctx, name); err != nil {
		return fmt.Errorf("error deleting host vSwitch: %s", err)
	}

	return nil
}
