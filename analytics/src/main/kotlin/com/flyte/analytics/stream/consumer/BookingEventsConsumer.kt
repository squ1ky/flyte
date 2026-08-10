package com.flyte.analytics.stream.consumer

import com.flyte.analytics.stream.model.BookingEventEnvelope
import org.slf4j.LoggerFactory
import org.springframework.kafka.annotation.KafkaListener
import org.springframework.stereotype.Component

@Component
class BookingEventsConsumer {

    private val log = LoggerFactory.getLogger(BookingEventsConsumer::class.java)

    @KafkaListener(
        topics = ["\${flyte.kafka.topics.booking-events}"],
        containerFactory = "bookingEventListenerContainerFactory"
    )
    fun listen(envelope: BookingEventEnvelope) {
        log.info(
            "booking event received: bookingId={}, eventType={}, payload={}",
            envelope.bookingId, envelope.eventType, envelope.payload
        )
    }
}