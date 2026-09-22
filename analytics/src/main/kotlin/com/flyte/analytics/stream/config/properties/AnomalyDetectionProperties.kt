package com.flyte.analytics.stream.config.properties

import org.springframework.boot.context.properties.ConfigurationProperties

@ConfigurationProperties(prefix = "flyte.analytics.anomaly")
data class AnomalyDetectionProperties(
    val cusum: Map<String, CusumParams> = emptyMap(),
)

data class CusumParams(
    val baseline: Double,
    val threshold: Double,
    val drift: Double,
)
