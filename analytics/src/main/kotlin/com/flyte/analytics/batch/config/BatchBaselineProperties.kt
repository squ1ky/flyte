package com.flyte.analytics.batch.config

import org.springframework.boot.context.properties.ConfigurationProperties
import java.time.Duration

@ConfigurationProperties(prefix = "flyte.analytics.batch")
data class BatchBaselineProperties(
    val enabled: Boolean = false,
    val pollInterval: Duration,
    val lookback: Duration,
    val bookingDb: BookingDbProperties,
)

data class BookingDbProperties(
    val url: String,
    val username: String,
    val password: String,
)
