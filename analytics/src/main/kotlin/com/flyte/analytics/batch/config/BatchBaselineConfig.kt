package com.flyte.analytics.batch.config

import com.flyte.analytics.anomaly.AnomalyApproach
import com.flyte.analytics.anomaly.AnomalyMetricNames
import com.flyte.analytics.anomaly.CusumMonitorFactory
import com.flyte.analytics.anomaly.MetricAnomalyMonitor
import com.flyte.analytics.batch.BatchBaselinePoller
import com.flyte.analytics.batch.repository.BookingCohortRepository
import com.flyte.analytics.common.AnalyticsWindowProperties
import com.zaxxer.hikari.HikariDataSource
import org.springframework.beans.factory.annotation.Qualifier
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty
import org.springframework.boot.context.properties.EnableConfigurationProperties
import org.springframework.boot.jdbc.DataSourceBuilder
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import javax.sql.DataSource

@Configuration
@ConditionalOnProperty(prefix = "flyte.analytics.batch", name = ["enabled"], havingValue = "true")
@EnableConfigurationProperties(BatchBaselineProperties::class, AnalyticsWindowProperties::class)
class BatchBaselineConfig(
    private val properties: BatchBaselineProperties,
    private val windowProperties: AnalyticsWindowProperties,
    private val monitorFactory: CusumMonitorFactory,
) {

    // defaultCandidate = false: the analytics-db DataSource must stay the only candidate for
    // autowiring by type, otherwise JdbcTemplate autoconfiguration (used by stream) backs off.
    @Bean(defaultCandidate = false)
    fun bookingDbDataSource(): DataSource =
        DataSourceBuilder.create()
            .type(HikariDataSource::class.java)
            .url(properties.bookingDb.url)
            .username(properties.bookingDb.username)
            .password(properties.bookingDb.password)
            .build()
            .apply {
                poolName = "booking-db-batch"
                maximumPoolSize = 2
            }

    @Bean
    fun batchConversionRateMonitor(): MetricAnomalyMonitor =
        monitorFactory.create(AnomalyMetricNames.CONVERSION_RATE, AnomalyApproach.BATCH)

    @Bean
    fun batchAbandonmentRateExpiredMonitor(): MetricAnomalyMonitor =
        monitorFactory.create(AnomalyMetricNames.ABANDONMENT_RATE_EXPIRED, AnomalyApproach.BATCH)

    @Bean(initMethod = "start", destroyMethod = "stop")
    fun batchBaselinePoller(
        @Qualifier("bookingDbDataSource") bookingDbDataSource: DataSource,
    ): BatchBaselinePoller =
        BatchBaselinePoller(
            repository = BookingCohortRepository(bookingDbDataSource),
            pollInterval = properties.pollInterval,
            lookback = properties.lookback,
            windowSizeSeconds = windowProperties.windowSize.seconds.toInt(),
            conversionRateMonitor = batchConversionRateMonitor(),
            abandonmentRateExpiredMonitor = batchAbandonmentRateExpiredMonitor(),
        )
}
