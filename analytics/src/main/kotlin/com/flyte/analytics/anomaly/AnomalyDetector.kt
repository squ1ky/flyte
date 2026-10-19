package com.flyte.analytics.anomaly

import java.time.Instant

interface AnomalyDetector {
    fun observe(value: Double, timestamp: Instant): AnomalyResult

    /**
     * Clears the alarm statistic before a new experiment run. What the detector has learned about
     * normal behaviour (if it learns anything, e.g. [PageHinkleyDetector]'s running mean) survives.
     */
    fun reset()
}

data class AnomalyResult(
    val isAnomaly: Boolean,
    val timestamp: Instant,
    val score: Double
)

/** Which shift is a degradation for the metric: a drop (success rates) or a rise (abandonment). */
enum class ChangeDirection { LOWER, UPPER }
