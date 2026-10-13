package com.flyte.analytics.common

import org.springframework.boot.context.properties.ConfigurationProperties
import java.time.Duration

@ConfigurationProperties(prefix = "flyte.analytics")
data class AnalyticsWindowProperties(
    val windowSize: Duration,
)