using Http;
using Rest;

namespace CallBooking;

@route("/api/bookings")
interface Bookings {
  @post create(@body booking: CreateBooking): { @statusCode _: 201; ...Booking } | ErrorResponse;
  @get list(): Booking[] | ErrorResponse;
  @delete @route("/{id}") delete(@path id: string): void | ErrorResponse;
}
