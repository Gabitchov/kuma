#include "extensions/filters/network/konvoy/config.h"

#include "extensions/filters/common/konvoy/anonymous_reporter.h"
#include "extensions/filters/common/konvoy/utility.h"
#include "extensions/filters/network/konvoy/konvoy.h"

#include "envoy/config/filter/network/konvoy/v2alpha/konvoy.pb.validate.h"

#include "envoy/registry/registry.h"

namespace Envoy {
namespace Extensions {
namespace NetworkFilters {
namespace Konvoy {

Network::FilterFactoryCb KonvoyFilterConfigFactory::createFilterFactoryFromProtoTyped(
    const envoy::config::filter::network::konvoy::v2alpha::Konvoy& proto_config,
    Server::Configuration::FactoryContext& context) {

  if (!anonymous_reporter_) {
    anonymous_reporter_ = Filters::Common::Konvoy::Utility::anonymousReporter(context);
  }
  anonymous_reporter_->observeUsageOfNetworkFilter();

  const auto filter_config =
      std::make_shared<Config>(proto_config, context.scope(), context.dispatcher().timeSource());
  Network::FilterFactoryCb callback;

  // gRPC client.
  callback = [grpc_service = proto_config.grpc_service(), &context,
              filter_config](Network::FilterManager& filter_manager) -> void {
    const auto async_client_factory =
        context.clusterManager().grpcAsyncClientManager().factoryForGrpcService(
            grpc_service, context.scope(), true);

    auto client = async_client_factory->create();

    filter_manager.addReadFilter(
        Network::ReadFilterSharedPtr{std::make_shared<Filter>(filter_config, std::move(client))});
  };

  return callback;
};

/**
 * Static registration for the Konvoy filter. @see RegisterFactory.
 */
REGISTER_FACTORY(KonvoyFilterConfigFactory, Server::Configuration::NamedNetworkFilterConfigFactory);

} // namespace Konvoy
} // namespace NetworkFilters
} // namespace Extensions
} // namespace Envoy
