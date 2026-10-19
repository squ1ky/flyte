package com.flyte.analytics.anomaly

import org.springframework.jdbc.core.JdbcTemplate
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class AnomalyMonitorFactoryTest {

    private val repository = AnomalyRepository(JdbcTemplate())

    private val metric = AnomalyMetricNames.CONVERSION_RATE

    private fun allParams(algorithms: List<String>) = AnomalyDetectionProperties(
        algorithms = algorithms,
        cusum = mapOf(metric to CusumParams(baseline = 0.8, threshold = 5.0, drift = 0.02)),
        ewma = mapOf(metric to EwmaParams(baseline = 0.8, sigma = 0.3, lambda = 0.1, width = 3.0)),
        pageHinkley = mapOf(metric to PageHinkleyParams(threshold = 5.0, delta = 0.02, alpha = 0.99, minObservations = 30)),
    )

    @Test
    fun `creates one monitor per enabled algorithm`() {
        val factory = AnomalyMonitorFactory(allParams(AnomalyAlgorithms.ALL), repository)

        val group = factory.create(metric, AnomalyApproach.STREAM)

        assertEquals(AnomalyAlgorithms.ALL, group.monitors.map { it.algorithm })
        assertEquals(setOf(metric), group.monitors.map { it.metricName }.toSet())
        assertEquals(setOf(AnomalyApproach.STREAM), group.monitors.map { it.approach }.toSet())
    }

    @Test
    fun `fails fast on missing parameters for an enabled algorithm`() {
        val factory = AnomalyMonitorFactory(allParams(AnomalyAlgorithms.ALL).copy(ewma = emptyMap()), repository)

        assertFailsWith<IllegalStateException> { factory.create(metric, AnomalyApproach.STREAM) }
    }

    @Test
    fun `rejects unknown algorithm names`() {
        assertFailsWith<IllegalArgumentException> {
            AnomalyMonitorFactory(allParams(listOf("cusum", "adwin")), repository)
        }
    }
}
