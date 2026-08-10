package com.flyte.analytics.stream.model

import com.fasterxml.jackson.annotation.JsonSubTypes
import com.fasterxml.jackson.annotation.JsonTypeInfo
import tools.jackson.databind.PropertyNamingStrategies
import tools.jackson.databind.annotation.JsonNaming

@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy::class)
data class BookingEventEnvelope(
    val bookingId: String,
    val eventType: BookingEventType,

    @JsonTypeInfo(
        use = JsonTypeInfo.Id.NAME,
        include = JsonTypeInfo.As.EXTERNAL_PROPERTY,
        property = "event_type"
    )
    @JsonSubTypes(
        JsonSubTypes.Type(value = BookingCreatedPayload::class, name = "booking_created"),
        JsonSubTypes.Type(value = BookingPaidPayload::class, name = "booking_paid"),
        JsonSubTypes.Type(value = BookingCancelledPayload::class, name = "booking_cancelled"),
    )
    val payload: BookingEventPayload
)