package com.flyte.analytics.stream.model

import tools.jackson.databind.PropertyNamingStrategies
import tools.jackson.databind.annotation.JsonNaming
import java.time.Instant

sealed interface BookingEventPayload

@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy::class)
data class BookingCreatedPayload(
    val bookingId: String,
    val userId: Long,
    val flightId: Long,
    val totalPriceCents: Long,
    val currency: String,
    val createdAt: Instant
) : BookingEventPayload

@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy::class)
data class BookingPaidPayload(
    val bookingId: String,
    val paidAt: Instant
) : BookingEventPayload

@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy::class)
data class BookingCancelledPayload(
    val bookingId: String,
    val reason: CancelReason,
    val cancelledAt: Instant
) : BookingEventPayload