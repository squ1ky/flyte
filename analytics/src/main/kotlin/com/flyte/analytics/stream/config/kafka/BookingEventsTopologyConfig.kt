package com.flyte.analytics.stream.config.kafka

import com.flyte.analytics.stream.StreamProcessorNames
import com.flyte.analytics.stream.StreamStoreNames
import com.flyte.analytics.stream.anomaly.MetricAnomalyMonitor
import com.flyte.analytics.stream.config.properties.AnalyticsWindowProperties
import com.flyte.analytics.stream.config.properties.BookingEventsConsumerKafkaProperties
import com.flyte.analytics.stream.model.db.BookingWindowState
import com.flyte.analytics.stream.model.kafka.BookingEventEnvelope
import com.flyte.analytics.stream.processor.BookingEventsWindowProcessor
import com.flyte.analytics.stream.repository.MetricsWindowsRepository
import jakarta.annotation.PostConstruct
import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.StreamsBuilder
import org.apache.kafka.streams.kstream.Consumed
import org.apache.kafka.streams.kstream.Named
import org.apache.kafka.streams.state.KeyValueStore
import org.apache.kafka.streams.state.StoreBuilder
import org.apache.kafka.streams.state.Stores
import org.springframework.beans.factory.annotation.Qualifier
import org.springframework.boot.context.properties.EnableConfigurationProperties
import org.springframework.context.annotation.Configuration
import org.springframework.kafka.annotation.EnableKafkaStreams
import org.springframework.kafka.support.serializer.JacksonJsonSerde

@Configuration
@EnableKafkaStreams
@EnableConfigurationProperties(AnalyticsWindowProperties::class)
class BookingEventsTopologyConfig(
    private val streamsBuilder: StreamsBuilder,
    private val windowProperties: AnalyticsWindowProperties,
    private val kafkaProperties: BookingEventsConsumerKafkaProperties,
    private val metricsWindowsRepository: MetricsWindowsRepository,
    @Qualifier("conversionRateMonitor") private val conversionRateMonitor: MetricAnomalyMonitor,
    @Qualifier("abandonmentRateExpiredMonitor") private val abandonmentRateExpiredMonitor: MetricAnomalyMonitor,
) {

    @PostConstruct
    fun buildTopology() {
        val stateStoreBuilder: StoreBuilder<KeyValueStore<String, BookingWindowState>> =
            Stores.keyValueStoreBuilder(
                Stores.persistentKeyValueStore(StreamStoreNames.BOOKING_WINDOW_STATE),
                Serdes.String(),
                JacksonJsonSerde(BookingWindowState::class.java)
            )
        streamsBuilder.addStateStore(stateStoreBuilder)

        val envelopeSerde = JacksonJsonSerde(BookingEventEnvelope::class.java)

        streamsBuilder
            .stream<String, BookingEventEnvelope>(
                kafkaProperties.topic,
                Consumed.with(Serdes.String(), envelopeSerde)
            )
            .process(
                {
                    BookingEventsWindowProcessor(
                        windowProperties.windowSize,
                        metricsWindowsRepository,
                        conversionRateMonitor,
                        abandonmentRateExpiredMonitor,
                    )
                },
                Named.`as`(StreamProcessorNames.BOOKING_EVENTS_WINDOW),
                StreamStoreNames.BOOKING_WINDOW_STATE
            )
    }
}