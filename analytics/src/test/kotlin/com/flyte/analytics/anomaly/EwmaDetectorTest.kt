package com.flyte.analytics.anomaly

import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class EwmaDetectorTest {

    private val now = Instant.parse("2026-01-01T00:00:00Z")

    // Round numbers for hand calculation: limit = 0.8 - 3 * 0.4 * sqrt(0.1 / 1.9) ~= 0.525.
    private fun paymentDetector() = EwmaDetector(baseline = 0.8, sigma = 0.4, lambda = 0.1, width = 3.0)

    @Test
    fun `does not fire when observed rate matches baseline`() {
        val detector = paymentDetector()

        repeat(200) { cycle ->
            listOf(1.0, 1.0, 1.0, 1.0, 0.0).forEach { value ->
                val result = detector.observe(value, now.plusSeconds(cycle.toLong()))
                assertFalse(result.isAnomaly, "false positive at cycle=$cycle value=$value score=${result.score}")
            }
        }
    }

    @Test
    fun `fires after a few consecutive failures, not on the first ones`() {
        val detector = paymentDetector()

        val firedAt = (1..20).first { detector.observe(0.0, now).isAnomaly }

        // z = 0.8 * 0.9^k drops below ~0.525 at k = 5: a short burst (k <= 4) is tolerated.
        assertEquals(5, firedAt)
    }

    @Test
    fun `score is the smoothed value`() {
        val detector = paymentDetector()

        val result = detector.observe(0.0, now)

        assertEquals(0.72, result.score, 1e-9)
    }

    @Test
    fun `reset puts the smoothed value back to baseline`() {
        val detector = paymentDetector()
        repeat(4) { detector.observe(0.0, now) }

        detector.reset()

        // Without the reset this fifth failure would fire.
        assertFalse(detector.observe(0.0, now).isAnomaly)
    }

    @Test
    fun `UPPER direction fires on a sustained rise and ignores values at baseline`() {
        val detector = EwmaDetector(baseline = 0.0, sigma = 0.1, lambda = 0.1, width = 3.0, direction = ChangeDirection.UPPER)

        repeat(100) { assertFalse(detector.observe(0.0, now).isAnomaly) }
        assertTrue((1..10).any { detector.observe(0.5, now).isAnomaly })
    }

    @Test
    fun `rejects invalid parameters`() {
        assertFailsWith<IllegalArgumentException> { EwmaDetector(0.8, sigma = 0.4, lambda = 0.0, width = 3.0) }
        assertFailsWith<IllegalArgumentException> { EwmaDetector(0.8, sigma = 0.0, lambda = 0.1, width = 3.0) }
    }
}
