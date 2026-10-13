package com.flyte.analytics.anomaly

import org.springframework.boot.context.properties.EnableConfigurationProperties
import org.springframework.stereotype.Component

/**
 * Builds a fresh [MetricAnomalyMonitor] with its own [CusumDetector] instance for a metric.
 *
 * Shared by stream and batch: both approaches must run the exact same algorithm with the
 * exact same parameters, only the way data is delivered to the detector differs. Each call
 * returns an independent detector, so stream and batch never share accumulated state.
 */
@Component
@EnableConfigurationProperties(AnomalyDetectionProperties::class)
class CusumMonitorFactory(
    private val properties: AnomalyDetectionProperties,
    private val anomalyRepository: AnomalyRepository,
) {

    fun create(metricName: String, approach: String): MetricAnomalyMonitor {
        val params = properties.cusum[metricName]
            ?: error("Missing flyte.analytics.anomaly.cusum.$metricName in configuration")
        // Direction is a property of the metric's meaning, not an experiment parameter, so it lives in code.
        val direction = when (metricName) {
            AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED -> CusumDirection.UPPER
            else -> CusumDirection.LOWER
        }
        val detector = CusumDetector(params.baseline, params.threshold, params.drift, direction)
        return MetricAnomalyMonitor(detector, anomalyRepository, metricName, ALGORITHM, approach)
    }

    companion object {
        const val ALGORITHM = "cusum"
    }
}
