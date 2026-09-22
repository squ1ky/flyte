package com.flyte.analytics.stream.consumer

import com.flyte.analytics.stream.WindowMath
import com.flyte.analytics.stream.anomaly.MetricAnomalyMonitor
import com.flyte.analytics.stream.config.properties.AnalyticsWindowProperties
import com.flyte.analytics.stream.model.kafka.PaymentResult
import com.flyte.analytics.stream.model.kafka.PaymentStatus
import com.flyte.analytics.stream.repository.PaymentErrorProfileRepository
import io.github.oshai.kotlinlogging.KotlinLogging
import org.springframework.beans.factory.annotation.Qualifier
import org.springframework.kafka.annotation.KafkaListener
import org.springframework.stereotype.Component

@Component
class PaymentResultsConsumer(
    private val paymentErrorProfileRepository: PaymentErrorProfileRepository,
    @Qualifier("paymentSuccessRateMonitor") private val paymentSuccessRateMonitor: MetricAnomalyMonitor,
    private val windowProperties: AnalyticsWindowProperties
) {

    private val log = KotlinLogging.logger {}

    private val windowSizeSeconds = windowProperties.windowSize.seconds.toInt()

    @KafkaListener(
        topics = ["\${flyte.kafka.consumer.payment-results.topic}"],
        containerFactory = "paymentResultListenerContainerFactory"
    )
    fun listen(result: PaymentResult) {
        val windowStart = WindowMath.truncateToWindow(result.processedAt, windowSizeSeconds)

        paymentSuccessRateMonitor.observe(
            value = if (result.status == PaymentStatus.SUCCESS) 1.0 else 0.0,
            timestamp = result.processedAt,
            windowStart = windowStart,
        )

        if (result.status != PaymentStatus.FAILED) return

        val errorMessage = result.errorMessage ?: run {
            log.warn { "payment_results FAILED without error_message: paymentId=${result.paymentId}" }
            return
        }

        paymentErrorProfileRepository.incrementError(windowStart, windowSizeSeconds, errorMessage)
    }
}