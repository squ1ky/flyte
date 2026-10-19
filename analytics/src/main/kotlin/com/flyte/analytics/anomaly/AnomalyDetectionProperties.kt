package com.flyte.analytics.anomaly

import org.springframework.boot.context.properties.ConfigurationProperties

/**
 * flyte.analytics.anomaly.*: which algorithms run, and per-metric parameters for each of them.
 * Every enabled algorithm must have parameters for every metric, otherwise startup fails.
 */
@ConfigurationProperties(prefix = "flyte.analytics.anomaly")
data class AnomalyDetectionProperties(
    val algorithms: List<String> = listOf(AnomalyAlgorithms.CUSUM),
    val cusum: Map<String, CusumParams> = emptyMap(),
    val ewma: Map<String, EwmaParams> = emptyMap(),
    val pageHinkley: Map<String, PageHinkleyParams> = emptyMap(),
)

data class CusumParams(
    val baseline: Double,
    val threshold: Double,
    val drift: Double,
)

data class EwmaParams(
    val baseline: Double,
    val sigma: Double,
    val lambda: Double,
    val width: Double,
)

data class PageHinkleyParams(
    val threshold: Double,
    val delta: Double,
    val alpha: Double,
    val minObservations: Int,
)
