package vsphere

import (
	"context"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
)

// hostStorageSystemFromHostSystemID locates a HostStorageSystem from a
// specified HostSystem managed object ID.
func hostStorageSystemFromHostSystemID(client *govmomi.Client, hsID string) (*object.HostStorageSystem, error) {
	hs, err := hostSystemFromID(client, hsID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	return hs.ConfigManager().StorageSystem(ctx)
}
