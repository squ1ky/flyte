package com.flyte.analytics.batch.repository

import com.flyte.analytics.batch.CohortCounts
import org.springframework.jdbc.core.JdbcTemplate
import java.sql.Timestamp
import java.time.Instant
import javax.sql.DataSource

/**
 * Reads booking outcomes straight from booking-db (the batch approach), grouped into the same
 * created_at cohorts the stream uses: window_start = floor(epoch(created_at) / size) * size,
 * which matches WindowMath.truncateToWindow.
 */
class BookingCohortRepository(
    bookingDbDataSource: DataSource,
) {
    private val jdbcTemplate = JdbcTemplate(bookingDbDataSource)

    fun fetchCohorts(since: Instant, windowSizeSeconds: Int): List<CohortCounts> =
        jdbcTemplate.query(
            """
            SELECT to_timestamp(floor(extract(epoch FROM created_at) / ?) * ?)              AS window_start,
                   COUNT(*)                                                                 AS created_count,
                   COUNT(*) FILTER (WHERE status = 'PAID')                                  AS paid_count,
                   COUNT(*) FILTER (WHERE status = 'CANCELLED' AND cancel_reason = 'expired')        AS cancelled_expired_count,
                   COUNT(*) FILTER (WHERE status = 'CANCELLED' AND cancel_reason = 'payment_failed') AS cancelled_payment_failed_count,
                   COUNT(*) FILTER (WHERE status = 'CANCELLED' AND cancel_reason = 'user_cancelled') AS cancelled_user_cancelled_count
            FROM bookings
            WHERE created_at >= ?
            GROUP BY 1
            ORDER BY 1
            """.trimIndent(),
            { rs, _ ->
                CohortCounts(
                    windowStart = rs.getTimestamp("window_start").toInstant(),
                    createdCount = rs.getInt("created_count"),
                    paidCount = rs.getInt("paid_count"),
                    cancelledExpiredCount = rs.getInt("cancelled_expired_count"),
                    cancelledPaymentFailedCount = rs.getInt("cancelled_payment_failed_count"),
                    cancelledUserCancelledCount = rs.getInt("cancelled_user_cancelled_count"),
                )
            },
            windowSizeSeconds, windowSizeSeconds, Timestamp.from(since)
        )
}
