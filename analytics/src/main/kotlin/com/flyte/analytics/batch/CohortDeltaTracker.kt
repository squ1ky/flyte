package com.flyte.analytics.batch

import com.flyte.analytics.anomaly.AnomalyMetricNames
import java.time.Instant

data class CohortCounts(
    val windowStart: Instant,
    val createdCount: Int,
    val paidCount: Int,
    val cancelledExpiredCount: Int,
    val cancelledPaymentFailedCount: Int,
    val cancelledUserCancelledCount: Int,
) {
    val terminalCount: Int
        get() = paidCount + cancelledExpiredCount + cancelledPaymentFailedCount + cancelledUserCancelledCount
}

data class CohortObservation(
    val metricName: String,
    val windowStart: Instant,
    val value: Double,
)

/**
 * Turns successive batch snapshots of cohort counts into detector observations.
 *
 * Stream feeds the detector once per terminal event (conversion_rate on every paid/cancelled,
 * abandonment_rate_expired on every expired). To keep the comparison about delivery latency only,
 * batch feeds the same number of observations: one per terminal outcome that appeared since the
 * previous snapshot, using the rate from the current snapshot. Otherwise a CUSUM threshold would
 * mean "N events" for stream and "N polls" for batch, and the two would not be comparable.
 *
 * The first snapshot only seeds the state: bookings that existed before the poller started are
 * history, not something batch has "just discovered".
 *
 * Not thread-safe: meant to be driven by a single poller thread.
 */
class CohortDeltaTracker {

    private var lastSeen: Map<Instant, CohortCounts> = emptyMap()
    private var seeded = false

    fun update(snapshot: List<CohortCounts>): List<CohortObservation> {
        val observations = if (seeded) snapshot.flatMap { observationsFor(it, lastSeen[it.windowStart]) } else emptyList()

        // Windows that left the lookback are simply dropped; the lookback only moves forward.
        lastSeen = snapshot.associateBy { it.windowStart }
        seeded = true
        return observations
    }

    private fun observationsFor(current: CohortCounts, previous: CohortCounts?): List<CohortObservation> {
        if (current.createdCount == 0) return emptyList()

        val newTerminal = current.terminalCount - (previous?.terminalCount ?: 0)
        val newExpired = current.cancelledExpiredCount - (previous?.cancelledExpiredCount ?: 0)

        val conversionRate = current.paidCount.toDouble() / current.createdCount
        val abandonmentRateExpired = current.cancelledExpiredCount.toDouble() / current.createdCount

        return List(maxOf(newTerminal, 0)) {
            CohortObservation(AnomalyMetricNames.CONVERSION_RATE, current.windowStart, conversionRate)
        } + List(maxOf(newExpired, 0)) {
            CohortObservation(AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED, current.windowStart, abandonmentRateExpired)
        }
    }
}
