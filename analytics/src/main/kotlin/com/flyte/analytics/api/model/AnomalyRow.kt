package com.flyte.analytics.api.model

import org.springframework.data.annotation.Id
import org.springframework.data.relational.core.mapping.Table
import java.time.Instant

@Table("anomalies")
data class AnomalyRow(
    @Id
    val id: Long,
    val algorithm: String,
    val approach: String,
    val metricName: String,
    val windowStart: Instant,
    val detectedAt: Instant,
    val score: Double,
    val isAnomaly: Boolean
)
