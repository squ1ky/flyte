package com.flyte.analytics.api.model

import java.time.Instant

data class WindowMetrics(
    val windowStart: Instant,
    val windowSizeSeconds: Int,
    val conversionRate: Double?,
    val abandonmentRateExpired: Double?,
    val abandonmentRatePaymentFailed: Double?,
    val abandonmentRateUserCancelled: Double?,
    val createdCount: Int
)