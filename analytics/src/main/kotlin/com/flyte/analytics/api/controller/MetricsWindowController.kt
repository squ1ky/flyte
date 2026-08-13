package com.flyte.analytics.api.controller

import com.flyte.analytics.api.model.WindowMetrics
import com.flyte.analytics.api.service.MetricsWindowService
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController
import reactor.core.publisher.Flux

@RestController
class MetricsWindowController(
    private val metricsWindowService: MetricsWindowService
) {

    @GetMapping("/api/metrics/windows")
    fun getRecentMetrics(
        @RequestParam(defaultValue = "60") windowSizeSeconds: Int,
        @RequestParam(defaultValue = "50") limit: Int
    ): Flux<WindowMetrics> = metricsWindowService.getRecentMetrics(windowSizeSeconds, limit)
}