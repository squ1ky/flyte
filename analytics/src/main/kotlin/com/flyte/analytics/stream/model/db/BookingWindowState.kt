package com.flyte.analytics.stream.model.db

import com.flyte.analytics.stream.model.kafka.CancelReason
import java.time.Instant


data class BookingWindowState(
    val bookingId: String,
    val windowStart: Instant,
    val status: BookingStatus,
    val cancelReason: CancelReason? = null
)

enum class BookingStatus {
    PENDING,
    PAID,
    CANCELLED,
}