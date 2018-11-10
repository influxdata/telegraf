package vsphere

import (
	"context"
	"reflect"
	"strings"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
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

func (f *Finder) Find(ctx context.Context, resType, path string, dst interface{}) error {
	p := strings.Split(path, "/")
	flt := make([]property.Filter, len(p)-1)
	for i := 1; i < len(p); i++ {
		flt[i-1] = property.Filter{"name": p[i]}
	}
	objs := make(map[string]types.ObjectContent)
	err := f.descend(ctx, f.client.Client.ServiceContent.RootFolder, resType, flt, 0, objs)
	if err != nil {
		return err
	}
	objectContentToTypedArray(objs, dst)
	return nil
}

func (f *Finder) descend(ctx context.Context, root types.ManagedObjectReference, resType string,
	parts []property.Filter, pos int, objs map[string]types.ObjectContent) error {

	// No more tokens to match?
	if pos >= len(parts) {
		return nil
	}

	// Get children
	ct, ok := childTypes[root.Reference().Type]
	if !ok {
		// We don't know how to handle children of this type. Stop descending.
		return nil
	}
	m := view.NewManager(f.client.Client.Client)
	defer m.Destroy(ctx)
	v, err := m.CreateContainerView(ctx, root, ct, false)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)
	var content []types.ObjectContent

	err = v.Retrieve(ctx, ct, []string{"name"}, &content)
	if err != nil {
		return err
	}
	for _, c := range content {
		if !parts[pos].MatchPropertyList(c.PropSet) {
			continue
		}

		if _, ok := objs[root.Reference().String()]; ok {
			continue
		}

		if c.Obj.Type == resType {
			// We found what we're looking for. Consider it a leaf and stop descending
			objs[c.Obj.String()] = c
			continue
		}

		// Deal with recursive wildcards (**)
		inc := 1 // Normally we advance one token.
		if parts[pos]["name"] == "**" {
			if pos >= len(parts)-1 {
				inc = 0 // Can't advance past last token, so keep descending the tree
			} else {
				// Lookahead to next token. If it matches this child, we are out of
				// the recursive wildcard handling and we can advance TWO tokens ahead, since
				// the token that ended the recursive wildcard mode is now consumed.
				if parts[pos+1].MatchPropertyList(c.PropSet) {
					if pos < len(parts)-2 {
						inc = 2
					} else {
						inc = 0
					}
				} else {
					// We didn't break out of recursicve wildcard mode yet, so stay on this token.
					inc = 0

				}
			}
		}
		err := f.descend(ctx, c.Obj, resType, parts, pos+inc, objs)
		if err != nil {
			return err
		}
	}
	return nil
}

func nameFromObjectContent(o types.ObjectContent) string {
	for _, p := range o.PropSet {
		if p.Name == "name" {
			return p.Val.(string)
		}
	}
	return "<unknown>"
}

func objectContentToTypedArray(objs map[string]types.ObjectContent, dst interface{}) error {
	rt := reflect.TypeOf(dst)
	if rt == nil || rt.Kind() != reflect.Ptr {
		panic("need pointer")
	}

	rv := reflect.ValueOf(dst).Elem()
	if !rv.CanSet() {
		panic("cannot set dst")
	}
	for _, p := range objs {
		v, err := mo.ObjectContentToType(p)
		if err != nil {
			return err
		}

		vt := reflect.TypeOf(v)

		if !rv.Type().AssignableTo(vt) {
			// For example: dst is []ManagedEntity, res is []HostSystem
			if field, ok := vt.FieldByName(rt.Elem().Elem().Name()); ok && field.Anonymous {
				rv.Set(reflect.Append(rv, reflect.ValueOf(v).FieldByIndex(field.Index)))
				continue
			}
		}

		rv.Set(reflect.Append(rv, reflect.ValueOf(v)))
	}
	return nil
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
