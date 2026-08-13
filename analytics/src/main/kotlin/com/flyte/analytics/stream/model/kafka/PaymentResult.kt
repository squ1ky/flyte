package com.flyte.analytics.stream.model.kafka

import tools.jackson.databind.PropertyNamingStrategies
import tools.jackson.databind.annotation.JsonNaming
import java.time.Instant

@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy::class)
data class PaymentResult(
    val bookingId: String,
    val paymentId: String,
    val status: PaymentStatus,
    val errorMessage: String? = null,
    val processedAt: Instant
)

enum class PaymentStatus {
    SUCCESS,
    FAILED,
}