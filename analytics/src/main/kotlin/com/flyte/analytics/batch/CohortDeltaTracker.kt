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
 * Stream feeds the detectors one 0/1 observation per terminal event: conversion_rate gets 1 for a
 * paid booking, abandonment_rate_expired gets 1 for an expired one, both get 0 for any other outcome.
 * Batch feeds exactly the same values: per window, as many 1s as new paid (expired) outcomes since
 * the previous snapshot and as many 0s as the other new outcomes. Both approaches thus deliver the same
 * multiset of observations to identical detectors; only the moment of delivery differs, which is what
 * the experiment measures. booking-db has no terminal timestamp, so the order of outcomes within one
 * window between two polls is unknown: 1s and 0s are spread evenly instead of grouped.
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
        val newTerminal = maxOf(current.terminalCount - (previous?.terminalCount ?: 0), 0)
        val newPaid = maxOf(current.paidCount - (previous?.paidCount ?: 0), 0).coerceAtMost(newTerminal)
        val newExpired = maxOf(current.cancelledExpiredCount - (previous?.cancelledExpiredCount ?: 0), 0)
            .coerceAtMost(newTerminal)

        return outcomes(newPaid, newTerminal).map {
            CohortObservation(AnomalyMetricNames.CONVERSION_RATE, current.windowStart, it)
        } + outcomes(newExpired, newTerminal).map {
            CohortObservation(AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED, current.windowStart, it)
        }
    }

    /** [total] 0/1 values with [ones] 1s spread evenly among them (Bresenham-style). */
    private fun outcomes(ones: Int, total: Int): List<Double> =
        List(total) { i -> if ((i + 1L) * ones / total > i.toLong() * ones / total) 1.0 else 0.0 }
}
