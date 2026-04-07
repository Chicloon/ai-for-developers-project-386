using Http;
using Rest;

namespace CallBooking;

@route("/api/availability-rules")
interface AvailabilityRules {
  @get list(): AvailabilityRule[] | ErrorResponse;
  @post create(@body rule: CreateAvailabilityRule): { @statusCode _: 201; ...AvailabilityRule } | ErrorResponse;
  @put @route("/{id}") update(@path id: string, @body rule: CreateAvailabilityRule): AvailabilityRule | ErrorResponse;
  @delete @route("/{id}") delete(@path id: string): void | ErrorResponse;
}
