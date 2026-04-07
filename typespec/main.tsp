import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";
import "@typespec/openapi3";

import "./models.tsp";
import "./availability-rules.tsp";
import "./blocked-days.tsp";
import "./slots.tsp";
import "./bookings.tsp";

using Http;
using Rest;

@service(#{ title: "Call Booking API" })
@server("http://localhost:8080", "Local development server")
namespace CallBooking;
