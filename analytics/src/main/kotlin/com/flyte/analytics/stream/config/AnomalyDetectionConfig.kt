package com.flyte.analytics.stream.config

import com.flyte.analytics.stream.anomaly.AnomalyMetricNames
import com.flyte.analytics.stream.anomaly.CusumDetector
import com.flyte.analytics.stream.anomaly.CusumDirection
import com.flyte.analytics.stream.anomaly.MetricAnomalyMonitor
import com.flyte.analytics.stream.config.properties.AnomalyDetectionProperties
import com.flyte.analytics.stream.repository.AnomalyRepository
import org.springframework.boot.context.properties.EnableConfigurationProperties
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration

@Configuration
@EnableConfigurationProperties(AnomalyDetectionProperties::class)
class AnomalyDetectionConfig(
    private val properties: AnomalyDetectionProperties,
    private val anomalyRepository: AnomalyRepository,
) {

    @Bean
    fun paymentSuccessRateMonitor(): MetricAnomalyMonitor =
        cusumMonitor(AnomalyMetricNames.PAYMENT_SUCCESS_RATE)

    @Bean
    fun conversionRateMonitor(): MetricAnomalyMonitor =
        cusumMonitor(AnomalyMetricNames.CONVERSION_RATE)

    @Bean
    fun abandonmentRateExpiredMonitor(): MetricAnomalyMonitor =
        cusumMonitor(AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED)

    private fun cusumMonitor(metricName: String): MetricAnomalyMonitor {
        val params = properties.cusum[metricName]
            ?: error("Missing flyte.analytics.anomaly.cusum.$metricName in configuration")
        val direction = when (metricName) {
            AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED -> CusumDirection.UPPER
            else -> CusumDirection.LOWER
        }
        val detector = CusumDetector(params.baseline, params.threshold, params.drift, direction)
        return MetricAnomalyMonitor(detector, anomalyRepository, metricName, ALGORITHM)
    }

    companion object {
        private const val ALGORITHM = "cusum"
    }
}
