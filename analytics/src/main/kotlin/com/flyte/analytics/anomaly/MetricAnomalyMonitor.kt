package com.flyte.analytics.anomaly

import io.github.oshai.kotlinlogging.KotlinLogging
import java.time.Instant
import java.util.concurrent.atomic.AtomicBoolean

object AnomalyApproach {
    const val STREAM = "stream"
    const val BATCH = "batch"
}

class MetricAnomalyMonitor(
    private val detector: AnomalyDetector,
    private val anomalyRepository: AnomalyRepository,
    private val metricName: String,
    private val algorithm: String,
    private val approach: String,
) {
    private val log = KotlinLogging.logger {}

    private val latched = AtomicBoolean(false)

    fun observe(value: Double, timestamp: Instant, windowStart: Instant) {
        val result = detector.observe(value, timestamp)

        if (!result.isAnomaly) return
        if (!latched.compareAndSet(false, true)) return

        log.warn {
            "Anomaly detected: metric=$metricName algorithm=$algorithm approach=$approach " +
                    "score=${result.score} at=${result.timestamp}"
        }

        anomalyRepository.insert(
            algorithm = algorithm,
            approach = approach,
            metricName = metricName,
            windowStart = windowStart,
            score = result.score,
        )
    }

    fun resetForNewRun() {
        detector.reset()
        latched.set(false)
    }
}
