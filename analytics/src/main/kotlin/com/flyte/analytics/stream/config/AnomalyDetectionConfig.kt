package com.flyte.analytics.stream.config

import com.flyte.analytics.anomaly.AnomalyApproach
import com.flyte.analytics.anomaly.AnomalyMetricNames
import com.flyte.analytics.anomaly.AnomalyMonitorFactory
import com.flyte.analytics.anomaly.MetricMonitorGroup
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration

@Configuration
class AnomalyDetectionConfig(
    private val monitorFactory: AnomalyMonitorFactory,
) {

    @Bean
    fun paymentSuccessRateMonitor(): MetricMonitorGroup =
        monitorFactory.create(AnomalyMetricNames.PAYMENT_SUCCESS_RATE, AnomalyApproach.STREAM)

    @Bean
    fun conversionRateMonitor(): MetricMonitorGroup =
        monitorFactory.create(AnomalyMetricNames.CONVERSION_RATE, AnomalyApproach.STREAM)

    @Bean
    fun abandonmentRateExpiredMonitor(): MetricMonitorGroup =
        monitorFactory.create(AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED, AnomalyApproach.STREAM)
}
