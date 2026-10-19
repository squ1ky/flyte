package com.flyte.analytics.stream.processor

import com.flyte.analytics.stream.StreamStoreNames
import com.flyte.analytics.common.WindowMath
import com.flyte.analytics.anomaly.MetricMonitorGroup
import com.flyte.analytics.stream.model.db.BookingStatus
import com.flyte.analytics.stream.model.db.BookingWindowState
import com.flyte.analytics.stream.model.kafka.BookingCancelledPayload
import com.flyte.analytics.stream.model.kafka.BookingCreatedPayload
import com.flyte.analytics.stream.model.kafka.BookingEventEnvelope
import com.flyte.analytics.stream.model.kafka.BookingEventType
import com.flyte.analytics.stream.model.kafka.BookingPaidPayload
import com.flyte.analytics.stream.model.kafka.CancelReason
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
    private val metricsWindowsRepository: MetricsWindowsRepository,
    private val conversionRateMonitor: MetricMonitorGroup,
    private val abandonmentRateExpiredMonitor: MetricMonitorGroup,
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

        val windowStart = WindowMath.truncateToWindow(payload.createdAt, windowSizeSeconds)
        store.put(envelope.bookingId, BookingWindowState(envelope.bookingId, windowStart, BookingStatus.PENDING))
        metricsWindowsRepository.incrementCreated(windowStart, windowSizeSeconds)
        log.info { "processed BOOKING_CREATED: bookingId=${envelope.bookingId}, windowStart=$windowStart" }
    }

    private fun handlePaid(envelope: BookingEventEnvelope) {
        val payload = envelope.payload as? BookingPaidPayload ?: run {
            log.error { "BOOKING_PAID with unexpected payload type: bookingId=${envelope.bookingId}" }
            return
        }

        val state = store.get(envelope.bookingId) ?: run {
            log.warn { "booking_paid for unknown bookingId=${envelope.bookingId} — no prior booking_created seen" }
            return
        }

        store.put(envelope.bookingId, state.copy(status = BookingStatus.PAID))
        metricsWindowsRepository.incrementPaid(state.windowStart, windowSizeSeconds)
        observeOutcome(paid = true, expired = false, payload.paidAt, state.windowStart)
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
        observeOutcome(paid = false, expired = payload.reason == CancelReason.EXPIRED, payload.cancelledAt, state.windowStart)
    }

    /**
     * Detectors see the outcome of each booking (1/0), not the window ratio paid/created: that ratio
     * is observed while the window still holds PENDING bookings, so its normal level depends on when it
     * is looked at (stream ~0.47, batch ~0.72 on a live baseline) and no shared detector parameters fit
     * both approaches. The mean of the 0/1 outcomes is the conversion (expiry share) among resolved
     * bookings, the same for stream and batch; batch feeds the same values (CohortDeltaTracker).
     * Both metrics are observed on every outcome, so a total payment outage (no booking_paid at all)
     * still reaches the detector and Page-Hinkley learns the normal expiry level (0).
     */
    private fun observeOutcome(paid: Boolean, expired: Boolean, timestamp: Instant, windowStart: Instant) {
        conversionRateMonitor.observe(if (paid) 1.0 else 0.0, timestamp, windowStart)
        abandonmentRateExpiredMonitor.observe(if (expired) 1.0 else 0.0, timestamp, windowStart)
    }
}