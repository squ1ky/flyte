package com.flyte.analytics.stream.repository

import com.flyte.analytics.stream.model.kafka.CancelReason
import org.springframework.jdbc.core.JdbcTemplate
import org.springframework.stereotype.Repository
import java.sql.Timestamp
import java.time.Instant

@Repository
class MetricsWindowsRepository(
    private val jdbcTemplate: JdbcTemplate
) {

    fun incrementCreated(windowStart: Instant, windowSizeSeconds: Int) =
        upsert(windowStart, windowSizeSeconds, "created_count")

    fun incrementPaid(windowStart: Instant, windowSizeSeconds: Int) =
        upsert(windowStart, windowSizeSeconds, "paid_count")

    fun incrementCancelled(windowStart: Instant, windowSizeSeconds: Int, reason: CancelReason) {
        val column = when (reason) {
            CancelReason.EXPIRED -> "cancelled_expired_count"
            CancelReason.PAYMENT_FAILED -> "cancelled_payment_failed_count"
            CancelReason.USER_CANCELLED -> "cancelled_user_cancelled_count"
        }
        upsert(windowStart, windowSizeSeconds, column)
    }

    private fun upsert(windowStart: Instant, windowSizeSeconds: Int, column: String) {
        jdbcTemplate.update(
            """
            INSERT INTO metrics_windows (window_start, window_size_seconds, $column, updated_at)
            VALUES (?, ?, 1, now())
            ON CONFLICT (window_start, window_size_seconds)
            DO UPDATE SET $column = metrics_windows.$column + 1, updated_at = now()
            """.trimIndent(),
            Timestamp.from(windowStart), windowSizeSeconds
        )
    }
}