package com.flyte.analytics.anomaly

import java.time.Instant

/**
 * All detectors watching one metric for one approach. Every algorithm gets the exact same
 * sequence of values, which is what makes their detection times comparable.
 */
class MetricMonitorGroup(
    val metricName: String,
    val approach: String,
    val monitors: List<MetricAnomalyMonitor>,
) {
    fun observe(value: Double, timestamp: Instant, windowStart: Instant) {
        monitors.forEach { it.observe(value, timestamp, windowStart) }
    }

    fun resetForNewRun() {
        monitors.forEach { it.resetForNewRun() }
    }
}
