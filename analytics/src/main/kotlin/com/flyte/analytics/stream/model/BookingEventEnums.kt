package com.flyte.analytics.stream.model

import tools.jackson.databind.EnumNamingStrategies
import tools.jackson.databind.annotation.EnumNaming

@EnumNaming(EnumNamingStrategies.SnakeCaseStrategy::class)
enum class BookingEventType {
    BOOKING_CREATED,
    BOOKING_PAID,
    BOOKING_CANCELLED,
}

@EnumNaming(EnumNamingStrategies.SnakeCaseStrategy::class)
enum class CancelReason {
    USER_CANCELLED,
    PAYMENT_FAILED,
    EXPIRED,
}