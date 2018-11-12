package vsphere

import (
	"context"
	"log"
	"reflect"
	"strings"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

var childTypes map[string][]string

var addFields map[string][]string

type Finder struct {
	client *Client
}

type ResourceFilter struct {
	finder  *Finder
	resType string
	paths   []string
}

type nameAndRef struct {
	name string
	ref  types.ManagedObjectReference
}

func (f *Finder) FindAll(ctx context.Context, resType string, paths []string, dst interface{}) error {
	for _, p := range paths {
		if err := f.Find(ctx, resType, p, dst); err != nil {
			return err
		}
	}
	return nil
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
	log.Printf("D! [input.vsphere] Find(%s, %s) returned %d objects", resType, path, len(objs))
	return nil
}

func (f *Finder) descend(ctx context.Context, root types.ManagedObjectReference, resType string,
	tokens []property.Filter, pos int, objs map[string]types.ObjectContent) error {
	isLeaf := pos == len(tokens)-1

	// No more tokens to match?
	if pos >= len(tokens) {
		return nil
	}

	// Determine child types

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

	// If we're at a potential leaf, we need to collect all properties specified for a target type. However,
	// if we're reached a node that may have multiple types of children, we have to do it in two
	// passes, since asking for fields that don't exist in all types will throw an error.
	// This is needed because of recursive wildcards. Even if we're at the last token, we can't determine
	// whether we've actually reached a leaf. This would happen for e.g. "/DC0/vm/**".
	fields := []string{"name"}
	if isLeaf {
		// Filter out the requested type from potential fields.
		fct := make([]string, 0, len(ct))
		for _, t := range ct {
			if t != resType {
				fct = append(fct, t)
			}
		}
		// Was the type present? (I.e. did we remove anything)
		if len(ct) != len(fct) {
			// Make a pass without the requested type with just the standard fields
			if len(fct) > 0 {
				err = v.Retrieve(ctx, fct, fields, &content)
				if err != nil {
					return err
				}
			}

			// Now make a pass with a full set of fields, but only for the requested type
			if af, ok := addFields[resType]; ok {
				fields = append(fields, af...)
			}
			var content1 []types.ObjectContent
			err = v.Retrieve(ctx, []string{resType}, fields, &content1)
			if err != nil {
				return err
			}
			content = append(content, content1...)
		} else {
			// The requested type wasn't part of potential children, so just collect the basics
			err = v.Retrieve(ctx, ct, fields, &content)

			if err != nil {
				return err
			}
		}
	} else {
		// Not at a leaf, so we can keep things simple
		err = v.Retrieve(ctx, ct, fields, &content)

		if err != nil {
			return err
		}
	}

	for _, c := range content {
		if !tokens[pos].MatchPropertyList(c.PropSet[:1]) {
			continue
		}

		// Already been here through another path? Skip!
		if _, ok := objs[root.Reference().String()]; ok {
			continue
		}

		if c.Obj.Type == resType && isLeaf {
			// We found what we're looking for. Consider it a leaf and stop descending
			objs[c.Obj.String()] = c
			continue
		}

		// Deal with recursive wildcards (**)
		inc := 1 // Normally we advance one token.
		if tokens[pos]["name"] == "**" {
			if isLeaf {
				inc = 0 // Can't advance past last token, so keep descending the tree
			} else {
				// Lookahead to next token. If it matches this child, we are out of
				// the recursive wildcard handling and we can advance TWO tokens ahead, since
				// the token that ended the recursive wildcard mode is now consumed.
				if tokens[pos+1].MatchPropertyList(c.PropSet) {
					if pos < len(tokens)-2 {
						inc = 2
					} else {
						// We found match and it's at a leaf! Grab it!
						objs[c.Obj.String()] = c
						continue
					}
				} else {
					// We didn't break out of recursicve wildcard mode yet, so stay on this token.
					inc = 0

				}
			}
		}
		err := f.descend(ctx, c.Obj, resType, tokens, pos+inc, objs)
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

func (r *ResourceFilter) FindAll(ctx context.Context, dst interface{}) error {
	return r.finder.FindAll(ctx, r.resType, r.paths, dst)
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

	addFields = map[string][]string{
		"HostSystem":             []string{"parent"},
		"VirtualMachine":         []string{"runtime.host", "config.guestId", "config.uuid"},
		"Datastore":              []string{"parent", "info"},
		"ClusterComputeResource": []string{"parent"},
		"Datacenter":             []string{"parent"},
	}
}
