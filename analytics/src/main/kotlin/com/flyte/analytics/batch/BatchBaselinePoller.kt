package com.flyte.analytics.batch

import com.flyte.analytics.anomaly.AnomalyMetricNames
import com.flyte.analytics.anomaly.MetricMonitorGroup
import com.flyte.analytics.batch.repository.BookingCohortRepository
import com.flyte.analytics.common.WindowMath
import io.github.oshai.kotlinlogging.KotlinLogging
import java.time.Duration
import java.time.Instant
import java.util.concurrent.Executors
import java.util.concurrent.ScheduledExecutorService
import java.util.concurrent.TimeUnit

/**
 * Batch baseline: every [pollInterval] queries booking-db for the cohort snapshot and feeds
 * the batch detectors. Detection latency of this approach is bounded below by the poll interval,
 * which is exactly what the experiment compares against the stream.
 */
class BatchBaselinePoller(
    private val repository: BookingCohortRepository,
    private val pollInterval: Duration,
    private val lookback: Duration,
    private val windowSizeSeconds: Int,
    private val conversionRateMonitor: MetricMonitorGroup,
    private val abandonmentRateExpiredMonitor: MetricMonitorGroup,
) {
    private val log = KotlinLogging.logger {}

    private val tracker = CohortDeltaTracker()
    private lateinit var executor: ScheduledExecutorService

    fun start() {
        executor = Executors.newSingleThreadScheduledExecutor { Thread(it, "batch-baseline-poller") }
        executor.scheduleWithFixedDelay(::pollSafely, 0, pollInterval.toMillis(), TimeUnit.MILLISECONDS)
        log.info { "batch baseline poller started: pollInterval=$pollInterval lookback=$lookback windowSize=${windowSizeSeconds}s" }
    }

    fun stop() {
        executor.shutdownNow()
    }

    private fun pollSafely() {
        // An exception escaping a scheduled task silently cancels all future runs.
        try {
            poll()
        } catch (e: Exception) {
            log.error(e) { "batch baseline poll failed, will retry in $pollInterval" }
        }
    }

    private fun poll() {
        val now = Instant.now()
        val since = WindowMath.truncateToWindow(now.minus(lookback), windowSizeSeconds)

        val snapshot = repository.fetchCohorts(since, windowSizeSeconds)
        val observations = tracker.update(snapshot)

        observations.forEach { observation ->
            val monitor = when (observation.metricName) {
                AnomalyMetricNames.CONVERSION_RATE -> conversionRateMonitor
                AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED -> abandonmentRateExpiredMonitor
                else -> error("unexpected batch metric ${observation.metricName}")
            }
            monitor.observe(observation.value, now, observation.windowStart)
        }

        log.debug { "batch poll: windows=${snapshot.size} observations=${observations.size}" }
    }
}
