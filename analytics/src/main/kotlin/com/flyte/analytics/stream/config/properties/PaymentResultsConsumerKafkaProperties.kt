package com.flyte.analytics.stream.config.properties

import org.springframework.boot.context.properties.ConfigurationProperties

@ConfigurationProperties(prefix = "flyte.kafka.consumer.payment-results")
data class PaymentResultsConsumerKafkaProperties(
    val topic: String,
)