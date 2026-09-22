package com.flyte.analytics.stream.repository

import com.flyte.analytics.stream.model.kafka.CancelReason
import org.springframework.jdbc.core.JdbcTemplate
import org.springframework.stereotype.Repository
import java.sql.Timestamp
import java.time.Instant

data class WindowCounts(
    val createdCount: Int,
    val paidCount: Int,
    val cancelledExpiredCount: Int,
    val cancelledPaymentFailedCount: Int,
    val cancelledUserCancelledCount: Int,
)

@Repository
class MetricsWindowsRepository(
    private val jdbcTemplate: JdbcTemplate
) {

    fun incrementCreated(windowStart: Instant, windowSizeSeconds: Int): WindowCounts =
        upsert(windowStart, windowSizeSeconds, "created_count")

    fun incrementPaid(windowStart: Instant, windowSizeSeconds: Int): WindowCounts =
        upsert(windowStart, windowSizeSeconds, "paid_count")

    fun incrementCancelled(windowStart: Instant, windowSizeSeconds: Int, reason: CancelReason): WindowCounts {
        val column = when (reason) {
            CancelReason.EXPIRED -> "cancelled_expired_count"
            CancelReason.PAYMENT_FAILED -> "cancelled_payment_failed_count"
            CancelReason.USER_CANCELLED -> "cancelled_user_cancelled_count"
        }
        return upsert(windowStart, windowSizeSeconds, column)
    }

    private fun upsert(windowStart: Instant, windowSizeSeconds: Int, column: String): WindowCounts =
        jdbcTemplate.queryForObject(
            """
            INSERT INTO metrics_windows (window_start, window_size_seconds, $column, updated_at)
            VALUES (?, ?, 1, now())
            ON CONFLICT (window_start, window_size_seconds)
            DO UPDATE SET $column = metrics_windows.$column + 1, updated_at = now()
            RETURNING created_count, paid_count, cancelled_expired_count,
                      cancelled_payment_failed_count, cancelled_user_cancelled_count
            """.trimIndent(),
            { rs, _ ->
                WindowCounts(
                    createdCount = rs.getInt("created_count"),
                    paidCount = rs.getInt("paid_count"),
                    cancelledExpiredCount = rs.getInt("cancelled_expired_count"),
                    cancelledPaymentFailedCount = rs.getInt("cancelled_payment_failed_count"),
                    cancelledUserCancelledCount = rs.getInt("cancelled_user_cancelled_count"),
                )
            },
            Timestamp.from(windowStart), windowSizeSeconds
        )!!
}