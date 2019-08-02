package rest

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

var From = &from{}

type from struct{}

func (c *from) Resource(r model.Resource) *Resource {
	return &Resource{
		Meta: ResourceMeta{
			Mesh: r.GetMeta().GetMesh(),
			Type: string(r.GetType()),
			Name: r.GetMeta().GetName(),
		},
		Spec: r.GetSpec(),
	}
}

func (c *from) ResourceList(rs model.ResourceList) *ResourceList {
	items := make([]*Resource, len(rs.GetItems()))
	for i, r := range rs.GetItems() {
		items[i] = c.Resource(r)
	}
	return &ResourceList{
		Items: items,
	}
}