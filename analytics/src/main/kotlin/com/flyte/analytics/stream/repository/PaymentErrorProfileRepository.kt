package com.flyte.analytics.stream.repository

import org.springframework.jdbc.core.JdbcTemplate
import org.springframework.stereotype.Repository
import java.sql.Timestamp
import java.time.Instant

@Repository
class PaymentErrorProfileRepository(
    private val jdbcTemplate: JdbcTemplate
) {

    fun incrementError(windowStart: Instant, windowSizeSeconds: Int, errorMessage: String) {
        jdbcTemplate.update(
            """
            INSERT INTO payment_error_profile (window_start, window_size_seconds, error_message, error_count, updated_at)
            VALUES (?, ?, ?, 1, now())
            ON CONFLICT (window_start, window_size_seconds, error_message)
            DO UPDATE SET error_count = payment_error_profile.error_count + 1, updated_at = now()
            """.trimIndent(),
            Timestamp.from(windowStart), windowSizeSeconds, errorMessage
        )
    }
}