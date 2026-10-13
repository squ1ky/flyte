package com.flyte.analytics.stream.config

import com.flyte.analytics.anomaly.AnomalyApproach
import com.flyte.analytics.anomaly.AnomalyMetricNames
import com.flyte.analytics.anomaly.CusumMonitorFactory
import com.flyte.analytics.anomaly.MetricAnomalyMonitor
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration

@Configuration
class AnomalyDetectionConfig(
    private val monitorFactory: CusumMonitorFactory,
) {

    @Bean
    fun paymentSuccessRateMonitor(): MetricAnomalyMonitor =
        monitorFactory.create(AnomalyMetricNames.PAYMENT_SUCCESS_RATE, AnomalyApproach.STREAM)

    @Bean
    fun conversionRateMonitor(): MetricAnomalyMonitor =
        monitorFactory.create(AnomalyMetricNames.CONVERSION_RATE, AnomalyApproach.STREAM)

    @Bean
    fun abandonmentRateExpiredMonitor(): MetricAnomalyMonitor =
        monitorFactory.create(AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED, AnomalyApproach.STREAM)
}
