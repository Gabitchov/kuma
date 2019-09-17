package server

import (
	"context"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"time"

	"github.com/Kong/kuma/pkg/core/permissions"
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/pkg/errors"

	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	util_watchdog "github.com/Kong/kuma/pkg/util/watchdog"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	xds_sync "github.com/Kong/kuma/pkg/xds/sync"
	xds_template "github.com/Kong/kuma/pkg/xds/template"
	xds_topology "github.com/Kong/kuma/pkg/xds/topology"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"

	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

func DefaultReconciler(rt core_runtime.Runtime) SnapshotReconciler {
	return &reconciler{
		&templateSnapshotGenerator{
			ProxyTemplateResolver: &simpleProxyTemplateResolver{
				ResourceManager:      manager.NewCachedManager(rt.ResourceManager(), 5*time.Second),
				DefaultProxyTemplate: xds_template.DefaultProxyTemplate,
			},
		},
		&simpleSnapshotCacher{rt.XDS().Hasher(), rt.XDS().Cache()},
	}
}

func DefaultDataplaneSyncTracker(rt core_runtime.Runtime, reconciler SnapshotReconciler) (envoy_xds.Callbacks, error) {
	permissionsMatcher := permissions.TrafficPermissionsMatcher{ResourceManager: rt.CachedResourceManager()}
	envoyCpCtx, err := xds_context.BuildControlPlaneContext(rt.Config())
	if err != nil {
		return nil, err
	}
	return xds_sync.NewDataplaneSyncTracker(func(key core_model.ResourceKey) util_watchdog.Watchdog {
		log := xdsServerLog.WithName("dataplane-sync-watchdog").WithValues("dataplaneKey", key)
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneConfigurationRefreshInterval)
			},
			OnTick: func() error {
				ctx := context.Background()
				dataplane := &mesh_core.DataplaneResource{}
				proxyID := xds.FromResourceKey(key)

				if err := rt.CachedResourceManager().Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
					if core_store.IsResourceNotFound(err) {
						return reconciler.Clear(&proxyID)
					}
					return err
				}

				meshList := mesh_core.MeshResourceList{}
				if err := rt.CachedResourceManager().List(ctx, &meshList, core_store.ListByMesh(proxyID.Mesh)); err != nil {
					return err
				}
				if len(meshList.Items) != 1 {
					return errors.Errorf("there should be a mesh of name %s. Found %d meshes of given name", proxyID.Mesh, len(meshList.Items))
				}
				envoyCtx := xds_context.Context{
					ControlPlane: envoyCpCtx,
					Mesh: xds_context.MeshContext{
						TlsEnabled:     meshList.Items[0].Spec.GetMtls().GetEnabled(),
						LoggingEnabled: meshList.Items[0].Spec.Logging.GetAccessLogs().GetEnabled(),
						LoggingPath:    meshList.Items[0].Spec.Logging.GetAccessLogs().GetFilePath(),
					},
				}

				outbound, err := xds_topology.GetOutboundTargets(ctx, dataplane, rt.CachedResourceManager())
				if err != nil {
					return err
				}

				matchedPermissions, err := permissionsMatcher.Match(ctx, dataplane)
				if err != nil {
					return err
				}

				proxy := xds.Proxy{
					Id:                 proxyID,
					Dataplane:          dataplane,
					TrafficPermissions: matchedPermissions,
					OutboundTargets:    outbound,
				}
				return reconciler.Reconcile(envoyCtx, &proxy)
			},
			OnError: func(err error) {
				log.Error(err, "OnTick() failed")
			},
		}
	}), nil
}

func DefaultDataplaneStatusTracker(rt core_runtime.Runtime) DataplaneStatusTracker {
	return NewDataplaneStatusTracker(rt, func(accessor SubscriptionStatusAccessor) DataplaneInsightSink {
		return NewDataplaneInsightSink(
			accessor,
			func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneStatusFlushInterval)
			},
			NewDataplaneInsightStore(rt.ResourceManager()))
	})
}
