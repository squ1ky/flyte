package com.flyte.analytics.stream.config.properties

import org.springframework.boot.context.properties.ConfigurationProperties

@ConfigurationProperties(prefix = "flyte.kafka.consumer.booking-events")
data class BookingEventsConsumerKafkaProperties(
    val topic: String,
)