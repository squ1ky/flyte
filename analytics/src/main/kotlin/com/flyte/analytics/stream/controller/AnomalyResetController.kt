package com.flyte.analytics.stream.controller

import com.flyte.analytics.stream.anomaly.MetricAnomalyMonitor
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController
import reactor.core.publisher.Mono

@RestController
@RequestMapping("/api/anomalies")
class AnomalyResetController(
    private val monitors: List<MetricAnomalyMonitor>,
) {

    @PostMapping("/reset")
    fun reset(): Mono<ResponseEntity<Void>> =
        Mono.fromCallable {
            monitors.forEach { it.resetForNewRun() }
            ResponseEntity.ok().build<Void>()
        }
}
