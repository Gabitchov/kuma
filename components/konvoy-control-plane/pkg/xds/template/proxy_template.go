package template

import (
	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
)

const (
	ProfileDefaultProxy = "default-proxy"
)

var (
	DefaultProxyTemplate = &konvoy_mesh.ProxyTemplate{
		Conf: []*konvoy_mesh.ProxyTemplateSource{
			&konvoy_mesh.ProxyTemplateSource{
				Type: &konvoy_mesh.ProxyTemplateSource_Profile{
					Profile: &konvoy_mesh.ProxyTemplateProfileSource{
						Name: ProfileDefaultProxy,
					},
				},
			},
		},
	}
)
