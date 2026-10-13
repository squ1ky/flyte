package com.flyte.analytics.batch

import com.flyte.analytics.anomaly.AnomalyMetricNames
import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class CohortDeltaTrackerTest {

    private val w1 = Instant.parse("2026-01-01T00:00:00Z")
    private val w2 = Instant.parse("2026-01-01T00:01:00Z")

    private fun counts(
        windowStart: Instant,
        created: Int,
        paid: Int = 0,
        expired: Int = 0,
        paymentFailed: Int = 0,
        userCancelled: Int = 0,
    ) = CohortCounts(windowStart, created, paid, expired, paymentFailed, userCancelled)

    @Test
    fun `first snapshot only seeds state`() {
        val tracker = CohortDeltaTracker()

        val observations = tracker.update(listOf(counts(w1, created = 10, paid = 8, expired = 2)))

        assertTrue(observations.isEmpty())
    }

    @Test
    fun `emits one conversion observation per new terminal outcome`() {
        val tracker = CohortDeltaTracker()
        tracker.update(listOf(counts(w1, created = 4, paid = 1)))

        val observations = tracker.update(listOf(counts(w1, created = 4, paid = 2, paymentFailed = 2)))

        val conversion = observations.filter { it.metricName == AnomalyMetricNames.CONVERSION_RATE }
        assertEquals(3, conversion.size)
        assertTrue(conversion.all { it.value == 0.5 && it.windowStart == w1 })
        assertTrue(observations.none { it.metricName == AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED })
    }

    @Test
    fun `emits abandonment observation per new expired, in addition to conversion`() {
        val tracker = CohortDeltaTracker()
        tracker.update(listOf(counts(w1, created = 5)))

        val observations = tracker.update(listOf(counts(w1, created = 5, expired = 2)))

        assertEquals(2, observations.count { it.metricName == AnomalyMetricNames.CONVERSION_RATE })
        val abandonment = observations.filter { it.metricName == AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED }
        assertEquals(2, abandonment.size)
        assertTrue(abandonment.all { it.value == 0.4 })
    }

    @Test
    fun `window appearing after seeding is treated as entirely new`() {
        val tracker = CohortDeltaTracker()
        tracker.update(listOf(counts(w1, created = 1, paid = 1)))

        val observations = tracker.update(
            listOf(counts(w1, created = 1, paid = 1), counts(w2, created = 3, paid = 1, paymentFailed = 1))
        )

        assertEquals(2, observations.size)
        assertTrue(observations.all { it.windowStart == w2 })
    }

    @Test
    fun `unchanged snapshot produces no observations`() {
        val tracker = CohortDeltaTracker()
        val snapshot = listOf(counts(w1, created = 3, paid = 2, expired = 1))
        tracker.update(snapshot)

        assertTrue(tracker.update(snapshot).isEmpty())
    }
}
