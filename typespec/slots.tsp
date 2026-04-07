using Http;
using Rest;

namespace CallBooking;

@route("/api/slots")
interface Slots {
  @get list(@query date: string): Slot[] | ErrorResponse;
}
