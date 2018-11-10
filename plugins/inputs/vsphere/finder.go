package vsphere

import (
	"context"
	"strings"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/types"
)

var childTypes map[string][]string

type Finder struct {
	client *Client
}

type nameAndRef struct {
	name string
	ref  types.ManagedObjectReference
}

func (f *Finder) Find(ctx context.Context, resType, path string) ([]types.ManagedObjectReference, error) {
	p := strings.Split(path, "/")
	flt := make([]property.Filter, len(p)-1)
	for i := 1; i < len(p); i++ {
		flt[i-1] = property.Filter{"name": p[i]}
	}
	moids, err := f.descend(ctx, f.client.Client.ServiceContent.RootFolder, resType, flt, 0)
	if err != nil {
		return nil, err
	}
	return moids, nil
}

func (f *Finder) descend(ctx context.Context, root types.ManagedObjectReference, resType string,
	parts []property.Filter, pos int) ([]types.ManagedObjectReference, error) {

	// We found what we're looking for. Consider it a leaf and stop descending
	if root.Reference().Type == resType {
		return []types.ManagedObjectReference{root}, nil
	}

	// No more tokens to match?
	if pos >= len(parts) {
		return []types.ManagedObjectReference{}, nil
	}

	// Get children
	ct, ok := childTypes[root.Reference().Type]
	if !ok {
		// We don't know how to handle children of this type. Stop descending.
		return []types.ManagedObjectReference{}, nil
	}
	m := view.NewManager(f.client.Client.Client)
	defer m.Destroy(ctx)
	v, err := m.CreateContainerView(ctx, root, ct, false)
	if err != nil {
		return nil, err
	}
	defer v.Destroy(ctx)
	var content []types.ObjectContent

	err = v.Retrieve(ctx, ct, []string{"name"}, &content)
	if err != nil {
		return nil, err
	}
	moids := make([]types.ManagedObjectReference, 0, 100)
	for _, c := range content {
		if !parts[pos].MatchPropertyList(c.PropSet) {
			continue
		}

		// Deal with recursive wildcards (**)
		inc := 1 // Normally we advance one token.
		if parts[pos]["name"] == "**" {
			if pos >= len(parts) {
				inc = 0 // Can't advance past last token, so keep descending the tree
			} else {
				// Lookahead to next token. If it matches this child, we are out of
				// the recursive wildcard handling and we can advance TWO tokens ahead, since
				// the token that ended the recursive wildcard mode is now consumed.
				if parts[pos+1].MatchPropertyList(c.PropSet) {
					if pos < len(parts)-3 {
						inc = 2
					} else {
						// We didn't break out of recursicve wildcard mode yet, so stay on this token.
						inc = 0
					}
				}
			}
		}
		r, err := f.descend(ctx, c.Obj, resType, parts, pos+inc)
		if err != nil {
			return nil, err
		}
		moids = append(moids, r...)
	}
	return moids, nil
}

func nameFromObjectContent(o types.ObjectContent) string {
	for _, p := range o.PropSet {
		if p.Name == "name" {
			return p.Val.(string)
		}
	}
	return "<unknown>"
}

func init() {
	childTypes = map[string][]string{
		"HostSystem":             []string{"VirtualMachine"},
		"ComputeResource":        []string{"HostSystem", "ResourcePool"},
		"ClusterComputeResource": []string{"HostSystem", "ResourcePool"},
		"Datacenter":             []string{"Folder"},
		"Folder": []string{
			"Folder",
			"Datacenter",
			"VirtualMachine",
			"ComputeResource",
			"ClusterComputeResource",
			"Datastore",
		},
	}
}
