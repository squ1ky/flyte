package com.flyte.analytics.stream.config.jdbc

import org.springframework.boot.jdbc.autoconfigure.DataSourceProperties
import org.springframework.boot.context.properties.EnableConfigurationProperties
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import javax.sql.DataSource

@Configuration
@EnableConfigurationProperties(DataSourceProperties::class)
class JdbcDataSourceConfig(
    private val dataSourceProperties: DataSourceProperties
) {

    @Bean
    fun dataSource(): DataSource =
        dataSourceProperties.initializeDataSourceBuilder().build()
}