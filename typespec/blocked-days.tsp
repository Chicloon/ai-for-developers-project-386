using Http;
using Rest;

namespace CallBooking;

@route("/api/blocked-days")
interface BlockedDays {
  @post create(@body day: CreateBlockedDay): { @statusCode _: 201; ...BlockedDay } | ErrorResponse;
  @delete @route("/{id}") delete(@path id: string): void | ErrorResponse;
}
