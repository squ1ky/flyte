package com.flyte.analytics.api.model

import org.springframework.data.annotation.Id
import org.springframework.data.relational.core.mapping.Table
import java.time.Instant

@Table("metrics_windows")
data class MetricsWindowRow(
    @Id
    val windowStart: Instant,
    val windowSizeSeconds: Int,
    val createdCount: Int,
    val paidCount: Int,
    val cancelledExpiredCount: Int,
    val cancelledPaymentFailedCount: Int,
    val cancelledUserCancelledCount: Int,
    val updatedAt: Instant
)