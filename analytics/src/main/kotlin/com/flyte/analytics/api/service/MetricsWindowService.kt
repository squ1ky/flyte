package com.flyte.analytics.api.service

import com.flyte.analytics.api.model.MetricsWindowRow
import com.flyte.analytics.api.model.WindowMetrics
import com.flyte.analytics.api.repository.MetricsWindowsReadRepository
import org.springframework.stereotype.Service
import reactor.core.publisher.Flux

@Service
class MetricsWindowService(
    private val repository: MetricsWindowsReadRepository
) {

    fun getRecentMetrics(windowSizeSeconds: Int, limit: Int): Flux<WindowMetrics> =
        repository.findRecent(windowSizeSeconds, limit).map { it.toWindowMetrics() }

    private fun MetricsWindowRow.toWindowMetrics(): WindowMetrics {
        val rate: (Int) -> Double? = { numerator ->
            if (createdCount == 0) null else numerator.toDouble() / createdCount
        }

        return WindowMetrics(
            windowStart = windowStart,
            windowSizeSeconds = windowSizeSeconds,
            conversionRate = rate(paidCount),
            abandonmentRateExpired = rate(cancelledExpiredCount),
            abandonmentRatePaymentFailed = rate(cancelledPaymentFailedCount),
            abandonmentRateUserCancelled = rate(cancelledUserCancelledCount),
            createdCount = createdCount
        )
    }
}