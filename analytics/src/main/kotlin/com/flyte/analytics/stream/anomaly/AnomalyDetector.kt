package com.flyte.analytics.stream.anomaly

import java.time.Instant

interface AnomalyDetector {
    fun observe(value: Double, timestamp: Instant): AnomalyResult
    fun reset()
}

data class AnomalyResult(
    val isAnomaly: Boolean,
    val timestamp: Instant,
    val score: Double
)