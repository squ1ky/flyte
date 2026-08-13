package com.flyte.analytics.stream.config

import com.flyte.analytics.stream.config.properties.BookingEventsConsumerKafkaProperties
import com.flyte.analytics.stream.model.kafka.BookingEventEnvelope
import org.apache.kafka.common.serialization.StringDeserializer
import org.springframework.boot.context.properties.EnableConfigurationProperties
import org.springframework.boot.kafka.autoconfigure.KafkaProperties
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.kafka.annotation.EnableKafka
import org.springframework.kafka.config.ConcurrentKafkaListenerContainerFactory
import org.springframework.kafka.core.ConsumerFactory
import org.springframework.kafka.core.DefaultKafkaConsumerFactory
import org.springframework.kafka.listener.DefaultErrorHandler
import org.springframework.kafka.support.serializer.JacksonJsonDeserializer

@Configuration
@EnableKafka
@EnableConfigurationProperties(BookingEventsConsumerKafkaProperties::class)
class BookingEventsConsumerKafkaConfig(
    private val kafkaProperties: KafkaProperties
) {

    @Bean
    fun bookingEventConsumerFactory(): ConsumerFactory<String, BookingEventEnvelope> {
        val valueDeserializer = JacksonJsonDeserializer(BookingEventEnvelope::class.java)
            .also { it.setUseTypeHeaders(false) }

        return DefaultKafkaConsumerFactory(
            kafkaProperties.buildConsumerProperties(),
            StringDeserializer(),
            valueDeserializer
        )
    }

    @Bean
    fun bookingEventListenerContainerFactory(
        consumerFactory: ConsumerFactory<String, BookingEventEnvelope>,
        errorHandler: DefaultErrorHandler,
    ): ConcurrentKafkaListenerContainerFactory<String, BookingEventEnvelope> {
        val factory = ConcurrentKafkaListenerContainerFactory<String, BookingEventEnvelope>()
        factory.setConsumerFactory(consumerFactory)
        factory.setCommonErrorHandler(errorHandler)
        return factory
    }
}