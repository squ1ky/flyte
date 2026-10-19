package com.flyte.analytics.anomaly

import org.springframework.boot.context.properties.EnableConfigurationProperties
import org.springframework.stereotype.Component

object AnomalyAlgorithms {
    const val CUSUM = "cusum"
    const val EWMA = "ewma"
    const val PAGE_HINKLEY = "page_hinkley"

    val ALL = listOf(CUSUM, EWMA, PAGE_HINKLEY)
}

/**
 * Builds a [MetricMonitorGroup] for a (metric, approach): one [MetricAnomalyMonitor] with its own
 * detector instance per enabled algorithm.
 *
 * Shared by stream and batch: both approaches must run the exact same algorithms with the
 * exact same parameters, only the way data is delivered to the detectors differs. Each call
 * returns independent detectors, so stream and batch never share accumulated state.
 */
@Component
@EnableConfigurationProperties(AnomalyDetectionProperties::class)
class AnomalyMonitorFactory(
    private val properties: AnomalyDetectionProperties,
    private val anomalyRepository: AnomalyRepository,
) {

    init {
        val unknown = properties.algorithms - AnomalyAlgorithms.ALL.toSet()
        require(unknown.isEmpty()) { "Unknown flyte.analytics.anomaly.algorithms $unknown, supported: ${AnomalyAlgorithms.ALL}" }
        require(properties.algorithms.isNotEmpty()) { "flyte.analytics.anomaly.algorithms is empty" }
    }

    fun create(metricName: String, approach: String): MetricMonitorGroup {
        val monitors = properties.algorithms.distinct().map { algorithm ->
            MetricAnomalyMonitor(createDetector(algorithm, metricName), anomalyRepository, metricName, algorithm, approach)
        }
        return MetricMonitorGroup(metricName, approach, monitors)
    }

    private fun createDetector(algorithm: String, metricName: String): AnomalyDetector {
        // Direction is a property of the metric's meaning, not an experiment parameter, so it lives in code.
        val direction = when (metricName) {
            AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED -> ChangeDirection.UPPER
            else -> ChangeDirection.LOWER
        }
        return when (algorithm) {
            AnomalyAlgorithms.CUSUM -> params(properties.cusum, "cusum", metricName).let {
                CusumDetector(it.baseline, it.threshold, it.drift, direction)
            }
            AnomalyAlgorithms.EWMA -> params(properties.ewma, "ewma", metricName).let {
                EwmaDetector(it.baseline, it.sigma, it.lambda, it.width, direction)
            }
            AnomalyAlgorithms.PAGE_HINKLEY -> params(properties.pageHinkley, "page-hinkley", metricName).let {
                PageHinkleyDetector(it.threshold, it.delta, it.alpha, it.minObservations, direction)
            }
            else -> error("unsupported algorithm $algorithm")
        }
    }

    private fun <T> params(byMetric: Map<String, T>, yamlKey: String, metricName: String): T =
        byMetric[metricName] ?: error("Missing flyte.analytics.anomaly.$yamlKey.$metricName in configuration")
}
