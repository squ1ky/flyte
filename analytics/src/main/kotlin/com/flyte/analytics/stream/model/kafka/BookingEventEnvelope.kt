package com.flyte.analytics.stream.model.kafka

import tools.jackson.core.JsonParser
import tools.jackson.databind.DeserializationContext
import tools.jackson.databind.JsonNode
import tools.jackson.databind.annotation.JsonDeserialize
import tools.jackson.databind.deser.std.StdDeserializer

@JsonDeserialize(using = BookingEventEnvelopeDeserializer::class)
data class BookingEventEnvelope(
    val bookingId: String,
    val eventType: BookingEventType,
    val payload: BookingEventPayload
)

class BookingEventEnvelopeDeserializer :
    StdDeserializer<BookingEventEnvelope>(BookingEventEnvelope::class.java) {

    override fun deserialize(p: JsonParser, ctxt: DeserializationContext): BookingEventEnvelope {
        val node: JsonNode = ctxt.readTree(p)

        val bookingId = node.get("booking_id")?.asString()
            ?: throw IllegalArgumentException("Missing 'booking_id' field in BookingEventEnvelope")
        val eventTypeRaw = node.get("event_type")?.asString()
            ?: throw IllegalArgumentException("Missing 'event_type' field in BookingEventEnvelope")
        val eventType = BookingEventType.valueOf(eventTypeRaw.uppercase())
        val payloadNode = node.get("payload")
            ?: throw IllegalArgumentException("Missing 'payload' field in BookingEventEnvelope")

        val payload: BookingEventPayload = when (eventType) {
            BookingEventType.BOOKING_CREATED -> ctxt.readTreeAsValue(payloadNode, BookingCreatedPayload::class.java)
            BookingEventType.BOOKING_PAID -> ctxt.readTreeAsValue(payloadNode, BookingPaidPayload::class.java)
            BookingEventType.BOOKING_CANCELLED -> ctxt.readTreeAsValue(payloadNode, BookingCancelledPayload::class.java)
        }

        return BookingEventEnvelope(bookingId, eventType, payload)
    }
}