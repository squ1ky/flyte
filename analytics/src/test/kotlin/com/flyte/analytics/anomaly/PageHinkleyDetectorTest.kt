package com.flyte.analytics.anomaly

import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class PageHinkleyDetectorTest {

    private val now = Instant.parse("2026-01-01T00:00:00Z")

    private fun paymentDetector() = PageHinkleyDetector(threshold = 5.0, delta = 0.01, alpha = 0.99, minObservations = 30)

    private fun PageHinkleyDetector.feedNormal(cycles: Int) = repeat(cycles) {
        listOf(1.0, 1.0, 1.0, 1.0, 0.0).forEach { value ->
            val result = observe(value, now)
            assertFalse(result.isAnomaly, "false positive on normal traffic, score=${result.score}")
        }
    }

    @Test
    fun `does not fire on stable traffic`() {
        paymentDetector().feedNormal(200)
    }

    @Test
    fun `fires on a sustained drop after learning the normal level`() {
        val detector = paymentDetector()
        detector.feedNormal(20)

        assertTrue((1..20).any { detector.observe(0.0, now).isAnomaly })
    }

    @Test
    fun `learns the level itself - no baseline to miscalibrate`() {
        // Normal level 0.6 instead of the expected 0.8: a CUSUM with baseline=0.8 false-positives here.
        val pageHinkley = paymentDetector()
        val cusum = CusumDetector(baseline = 0.8, threshold = 5.0, drift = 0.01)
        var cusumFired = false

        repeat(200) {
            listOf(1.0, 1.0, 1.0, 0.0, 0.0).forEach { value ->
                assertFalse(pageHinkley.observe(value, now).isAnomaly)
                cusumFired = cusumFired || cusum.observe(value, now).isAnomaly
            }
        }
        assertTrue(cusumFired)
    }

    @Test
    fun `stays silent until minObservations`() {
        val detector = paymentDetector()

        repeat(29) { assertFalse(detector.observe(0.0, now).isAnomaly) }
    }

    @Test
    fun `reset clears the statistic but keeps the learned mean`() {
        val detector = paymentDetector()
        detector.feedNormal(20)
        repeat(3) { detector.observe(0.0, now) }

        detector.reset()

        // Learned mean ~0.8 survives the reset, so a fault right after it is still detected.
        // Had the mean been cleared, it would be learned from the zeros and never fire.
        assertTrue((1..20).any { detector.observe(0.0, now).isAnomaly })
    }

    @Test
    fun `UPPER direction fires on a rise from a learned zero level`() {
        val detector = PageHinkleyDetector(threshold = 3.0, delta = 0.01, alpha = 0.99, minObservations = 30,
            direction = ChangeDirection.UPPER)
        repeat(100) { assertFalse(detector.observe(0.0, now).isAnomaly) }

        assertTrue((1..10).any { detector.observe(1.0, now).isAnomaly })
    }
}
