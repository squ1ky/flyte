package com.flyte.analytics.stream.controller

import com.flyte.analytics.anomaly.MetricMonitorGroup
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController
import reactor.core.publisher.Flux
import reactor.core.publisher.Mono

data class MonitorInfo(
    val metricName: String,
    val algorithm: String,
    val approach: String,
)

@RestController
@RequestMapping("/api/anomalies")
class AnomalyMonitorsController(
    groups: List<MetricMonitorGroup>,
) {

    private val monitors = groups.flatMap { it.monitors }

    /** Registered detectors (metric x algorithm x approach). The experiment runner waits for exactly this set. */
    @GetMapping("/monitors")
    fun list(): Flux<MonitorInfo> =
        Flux.fromIterable(monitors.map { MonitorInfo(it.metricName, it.algorithm, it.approach) })

    @PostMapping("/reset")
    fun reset(): Mono<ResponseEntity<Void>> =
        Mono.fromCallable {
            monitors.forEach { it.resetForNewRun() }
            ResponseEntity.ok().build<Void>()
        }
}
