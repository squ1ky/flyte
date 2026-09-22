package com.flyte.analytics.stream.repository

import org.springframework.jdbc.core.JdbcTemplate
import org.springframework.stereotype.Repository
import java.sql.Timestamp
import java.time.Instant

@Repository
class AnomalyRepository(
    private val jdbcTemplate: JdbcTemplate
) {

    fun insert(
        algorithm: String,
        approach: String,
        metricName: String,
        windowStart: Instant,
        score: Double,
    ) {
        jdbcTemplate.update(
            """
            INSERT INTO anomalies (algorithm, approach, metric_name, window_start, detected_at, score, is_anomaly)
            VALUES (?, ?, ?, ?, now(), ?, true)
            """.trimIndent(),
            algorithm, approach, metricName, Timestamp.from(windowStart), score
        )
    }
}