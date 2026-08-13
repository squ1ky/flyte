package com.flyte.analytics.stream.consumer

import com.flyte.analytics.stream.model.kafka.BookingEventEnvelope
import io.github.oshai.kotlinlogging.KotlinLogging
import org.springframework.kafka.annotation.KafkaListener
import org.springframework.stereotype.Component

@Component
class BookingEventsConsumer {

    private val log = KotlinLogging.logger {}

    @KafkaListener(
        topics = ["\${flyte.kafka.consumer.booking-events.topic}"],
        containerFactory = "bookingEventListenerContainerFactory"
    )
    fun listen(envelope: BookingEventEnvelope) {
        log.info {
            "booking event received: bookingId=${envelope.bookingId}, eventType=${envelope.eventType}, payload=${envelope.payload}"
        }
    }
}