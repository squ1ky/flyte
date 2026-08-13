package com.flyte.analytics.stream.processor

import com.flyte.analytics.stream.StreamStoreNames
import com.flyte.analytics.stream.model.db.BookingStatus
import com.flyte.analytics.stream.model.db.BookingWindowState
import com.flyte.analytics.stream.model.kafka.BookingCancelledPayload
import com.flyte.analytics.stream.model.kafka.BookingCreatedPayload
import com.flyte.analytics.stream.model.kafka.BookingEventEnvelope
import com.flyte.analytics.stream.model.kafka.BookingEventType
import com.flyte.analytics.stream.repository.MetricsWindowsRepository
import io.github.oshai.kotlinlogging.KLogger
import io.github.oshai.kotlinlogging.KotlinLogging
import org.apache.kafka.streams.processor.api.ContextualProcessor
import org.apache.kafka.streams.processor.api.ProcessorContext
import org.apache.kafka.streams.processor.api.Record
import org.apache.kafka.streams.state.KeyValueStore
import java.time.Duration
import java.time.Instant

class BookingEventsWindowProcessor(
    private val windowSize: Duration,
    private val metricsWindowsRepository: MetricsWindowsRepository
) : ContextualProcessor<String, BookingEventEnvelope, Void, Void>() {

    private val log: KLogger = KotlinLogging.logger {}

    private val windowSizeSeconds = windowSize.seconds.toInt()
    private lateinit var store: KeyValueStore<String, BookingWindowState>

    override fun init(context: ProcessorContext<Void, Void>) {
        super.init(context)
        store = context.getStateStore(StreamStoreNames.BOOKING_WINDOW_STATE)
    }

    override fun process(record: Record<String, BookingEventEnvelope>) {
        val envelope = record.value()
        when (envelope.eventType) {
            BookingEventType.BOOKING_CREATED -> handleCreated(envelope)
            BookingEventType.BOOKING_PAID -> handlePaid(envelope)
            BookingEventType.BOOKING_CANCELLED -> handleCancelled(envelope)
        }
    }

    private fun handleCreated(envelope: BookingEventEnvelope) {
        val payload = envelope.payload as? BookingCreatedPayload ?: run {
            log.error { "BOOKING_CREATED with unexpected payload type: bookingId=${envelope.bookingId}" }
            return
        }

        val windowStart = truncateToWindow(payload.createdAt)
        store.put(envelope.bookingId, BookingWindowState(envelope.bookingId, windowStart, BookingStatus.PENDING))
        metricsWindowsRepository.incrementCreated(windowStart, windowSizeSeconds)
    }

    private fun handlePaid(envelope: BookingEventEnvelope) {
        val state = store.get(envelope.bookingId) ?: run {
            log.warn { "booking_paid for unknown bookingId=${envelope.bookingId} — no prior booking_created seen" }
            return
        }

        store.put(envelope.bookingId, state.copy(status = BookingStatus.PAID))
        metricsWindowsRepository.incrementPaid(state.windowStart, windowSizeSeconds)
    }

    private fun handleCancelled(envelope: BookingEventEnvelope) {
        val payload = envelope.payload as? BookingCancelledPayload ?: run {
            log.error { "BOOKING_CANCELLED with unexpected payload type: bookingId=${envelope.bookingId}" }
            return
        }

        val state = store.get(envelope.bookingId) ?: run {
            log.warn { "booking_cancelled for unknown bookingId=${envelope.bookingId} — no prior booking_created seen" }
            return
        }

        store.put(envelope.bookingId, state.copy(status = BookingStatus.CANCELLED, cancelReason = payload.reason))
        metricsWindowsRepository.incrementCancelled(state.windowStart, windowSizeSeconds, payload.reason)
    }

    private fun truncateToWindow(instant: Instant): Instant {
        val epochSecond = instant.epochSecond
        val truncated = epochSecond - (epochSecond % windowSizeSeconds)
        return Instant.ofEpochSecond(truncated)
    }
}