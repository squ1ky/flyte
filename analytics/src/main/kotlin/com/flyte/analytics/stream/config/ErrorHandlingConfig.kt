package com.flyte.analytics.stream.config

import io.github.oshai.kotlinlogging.KotlinLogging
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.kafka.listener.ConsumerRecordRecoverer
import org.springframework.kafka.listener.DefaultErrorHandler
import org.springframework.util.backoff.FixedBackOff

@Configuration
class ErrorHandlingConfig {

    private val log = KotlinLogging.logger {}

    @Bean
    fun kafkaErrorHandler(): DefaultErrorHandler {
        val recoverer = ConsumerRecordRecoverer { record, exception ->
            log.error(exception) {
                "Skipping unprocessable record: topic=${record.topic()}, partition=${record.partition()}, offset=${record.offset()}, key=${record.key()}"
            }
        }
        return DefaultErrorHandler(recoverer, FixedBackOff(0L, 0L))
    }
}