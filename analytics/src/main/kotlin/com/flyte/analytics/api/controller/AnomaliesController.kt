package com.flyte.analytics.api.controller

import com.flyte.analytics.api.model.AnomalyRow
import com.flyte.analytics.api.repository.AnomaliesReadRepository
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController
import reactor.core.publisher.Flux

@RestController
class AnomaliesController(
    private val repository: AnomaliesReadRepository
) {

    @GetMapping("/api/anomalies")
    fun getRecentAnomalies(
        @RequestParam(required = false) metricName: String?,
        @RequestParam(required = false) algorithm: String?,
        @RequestParam(required = false) approach: String?,
        @RequestParam(defaultValue = "50") limit: Int
    ): Flux<AnomalyRow> = repository.findRecent(metricName, algorithm, approach, limit)
}
