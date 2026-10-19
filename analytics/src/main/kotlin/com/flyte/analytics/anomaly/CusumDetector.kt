package com.flyte.analytics.anomaly

import java.time.Instant

class CusumDetector(
    private val baseline: Double,
    private val threshold: Double,
    private val drift: Double,
    private val direction: ChangeDirection = ChangeDirection.LOWER,
) : AnomalyDetector {

    private var cumSum = 0.0

    @Synchronized
    override fun observe(value: Double, timestamp: Instant): AnomalyResult {
        val deviation = when (direction) {
            ChangeDirection.LOWER -> baseline - value
            ChangeDirection.UPPER -> value - baseline
        }
        cumSum = maxOf(0.0, cumSum + deviation - drift)
        return AnomalyResult(cumSum > threshold, timestamp, cumSum)
    }

    @Synchronized
    override fun reset() {
        cumSum = 0.0
    }
}
