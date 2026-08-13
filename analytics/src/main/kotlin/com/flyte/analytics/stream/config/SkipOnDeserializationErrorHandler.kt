package com.flyte.analytics.stream.config

import io.github.oshai.kotlinlogging.KotlinLogging
import org.apache.kafka.clients.consumer.ConsumerRecord
import org.apache.kafka.streams.errors.DeserializationExceptionHandler
import org.apache.kafka.streams.errors.ErrorHandlerContext

class SkipOnDeserializationErrorHandler : DeserializationExceptionHandler {

    private val log = KotlinLogging.logger {}

    override fun handleError(
        context: ErrorHandlerContext,
        record: ConsumerRecord<ByteArray, ByteArray>,
        exception: Exception
    ): DeserializationExceptionHandler.Response {
        log.error(exception) {
            "Skipping unprocessable record in Kafka Streams: topic=${record.topic()}, partition=${record.partition()}, offset=${record.offset()}"
        }
        return DeserializationExceptionHandler.Response.resume()
    }

    override fun configure(configs: MutableMap<String, *>?) {}
}