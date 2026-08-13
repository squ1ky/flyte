package com.flyte.analytics.stream.config

import com.flyte.analytics.stream.config.properties.PaymentResultsConsumerKafkaProperties
import com.flyte.analytics.stream.model.kafka.PaymentResult
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
@EnableConfigurationProperties(PaymentResultsConsumerKafkaProperties::class)
class PaymentResultsConsumerKafkaConfig(
    private val kafkaProperties: KafkaProperties
) {

    @Bean
    fun paymentResultConsumerFactory(): ConsumerFactory<String, PaymentResult> {
        val valueDeserializer = JacksonJsonDeserializer(PaymentResult::class.java)
            .also { it.setUseTypeHeaders(false) }

        return DefaultKafkaConsumerFactory(
            kafkaProperties.buildConsumerProperties(),
            StringDeserializer(),
            valueDeserializer
        )
    }

    @Bean
    fun paymentResultListenerContainerFactory(
        consumerFactory: ConsumerFactory<String, PaymentResult>,
        errorHandler: DefaultErrorHandler
    ): ConcurrentKafkaListenerContainerFactory<String, PaymentResult> {
        val factory = ConcurrentKafkaListenerContainerFactory<String, PaymentResult>()
        factory.setConsumerFactory(consumerFactory)
        factory.setCommonErrorHandler(errorHandler)
        return factory
    }
}