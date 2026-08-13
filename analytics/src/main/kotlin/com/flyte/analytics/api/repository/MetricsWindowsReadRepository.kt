package com.flyte.analytics.api.repository

import com.flyte.analytics.api.model.MetricsWindowRow
import org.springframework.data.domain.Sort
import org.springframework.data.r2dbc.core.R2dbcEntityTemplate
import org.springframework.data.relational.core.query.Criteria
import org.springframework.data.relational.core.query.Query
import org.springframework.stereotype.Repository
import reactor.core.publisher.Flux

@Repository
class MetricsWindowsReadRepository(
    private val template: R2dbcEntityTemplate
) {

    fun findRecent(windowSizeSeconds: Int, limit: Int): Flux<MetricsWindowRow> {
        val query = Query.query(Criteria.where("window_size_seconds").`is`(windowSizeSeconds))
            .sort(Sort.by(Sort.Direction.DESC, "window_start"))
            .limit(limit)

        return template.select(MetricsWindowRow::class.java).matching(query).all()
    }
}