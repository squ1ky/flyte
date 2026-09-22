package com.flyte.analytics.stream.anomaly

import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class CusumDetectorTest {

    private val now = Instant.parse("2026-01-01T00:00:00Z")

    @Test
    fun `does not fire when observed rate matches baseline`() {
        val detector = CusumDetector(baseline = 0.8, threshold = 5.0, drift = 0.01)

        repeat(200) { cycle ->
            listOf(1.0, 1.0, 1.0, 1.0, 0.0).forEach { value ->
                val result = detector.observe(value, now.plusSeconds(cycle.toLong()))
                assertFalse(result.isAnomaly, "false positive at cycle=$cycle value=$value score=${result.score}")
            }
        }
    }

    @Test
    fun `fires when observed rate drops well below baseline`() {
        val detector = CusumDetector(baseline = 0.8, threshold = 5.0, drift = 0.01)

        val results = (1..20).map { i -> detector.observe(0.0, now.plusSeconds(i.toLong())) }

        assertTrue(results.any { it.isAnomaly }, "detector never fired under sustained total failure")
    }

    @Test
    fun `reset clears accumulated score`() {
        val detector = CusumDetector(baseline = 0.8, threshold = 5.0, drift = 0.01)
        repeat(19) { detector.observe(0.0, now) }
        assertTrue(detector.observe(0.0, now).isAnomaly)

        detector.reset()

        assertFalse(detector.observe(1.0, now).isAnomaly)
    }

    @Test
    fun `UPPER direction does not fire while value stays at or below baseline`() {
        val detector = CusumDetector(baseline = 0.0, threshold = 3.0, drift = 0.01, direction = CusumDirection.UPPER)

        repeat(50) {
            val result = detector.observe(0.0, now)
            assertFalse(result.isAnomaly, "false positive at value=baseline, score=${result.score}")
        }
    }

    @Test
    fun `UPPER direction fires when value sustains above baseline`() {
        val detector = CusumDetector(baseline = 0.0, threshold = 3.0, drift = 0.01, direction = CusumDirection.UPPER)

        val results = (1..10).map { i -> detector.observe(1.0, now.plusSeconds(i.toLong())) }

        assertTrue(results.any { it.isAnomaly }, "detector never fired under sustained abandonment_rate_expired=1.0")
    }

    @Test
    fun `baseline that does not match the true rate eventually false-positives`() {
        val detector = CusumDetector(baseline = 0.95, threshold = 5.0, drift = 0.01)

        val results = (1..500).map { i ->
            val value = if (i % 5 == 0) 0.0 else 1.0
            detector.observe(value, now.plusSeconds(i.toLong()))
        }

        assertTrue(results.any { it.isAnomaly }, "expected miscalibrated baseline to eventually false-positive")
    }
}
