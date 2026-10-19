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

    private fun List<CohortObservation>.values(metric: String) = filter { it.metricName == metric }.map { it.value }

    @Test
    fun `emits one 0-1 outcome per metric per new terminal outcome, like stream`() {
        val tracker = CohortDeltaTracker()
        tracker.update(listOf(counts(w1, created = 4, paid = 1)))

        val observations = tracker.update(listOf(counts(w1, created = 4, paid = 2, paymentFailed = 2)))

        // 3 new outcomes: 1 paid, 2 payment_failed.
        assertEquals(listOf(0.0, 1.0, 0.0).sorted(), observations.values(AnomalyMetricNames.CONVERSION_RATE).sorted())
        assertTrue(observations.all { it.windowStart == w1 })
        // No expirations: abandonment is still observed (at its normal level, 0) so detectors learn it.
        assertEquals(listOf(0.0, 0.0, 0.0), observations.values(AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED))
    }

    @Test
    fun `expired outcomes are 1 for abandonment and 0 for conversion`() {
        val tracker = CohortDeltaTracker()
        tracker.update(listOf(counts(w1, created = 5)))

        val observations = tracker.update(listOf(counts(w1, created = 5, expired = 2)))

        assertEquals(listOf(0.0, 0.0), observations.values(AnomalyMetricNames.CONVERSION_RATE))
        assertEquals(listOf(1.0, 1.0), observations.values(AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED))
    }

    @Test
    fun `1s and 0s of one poll are spread evenly, not grouped`() {
        val tracker = CohortDeltaTracker()
        tracker.update(listOf(counts(w1, created = 10)))

        val observations = tracker.update(listOf(counts(w1, created = 10, paid = 5, paymentFailed = 5)))

        assertEquals(
            listOf(0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0),
            observations.values(AnomalyMetricNames.CONVERSION_RATE),
        )
    }

    @Test
    fun `window appearing after seeding is treated as entirely new`() {
        val tracker = CohortDeltaTracker()
        tracker.update(listOf(counts(w1, created = 1, paid = 1)))

        val observations = tracker.update(
            listOf(counts(w1, created = 1, paid = 1), counts(w2, created = 3, paid = 1, paymentFailed = 1))
        )

        assertEquals(4, observations.size)
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
