package com.flyte.analytics.api.repository

import com.flyte.analytics.api.model.AnomalyRow
import org.springframework.data.domain.Sort
import org.springframework.data.r2dbc.core.R2dbcEntityTemplate
import org.springframework.data.relational.core.query.Criteria
import org.springframework.data.relational.core.query.Query
import org.springframework.stereotype.Repository
import reactor.core.publisher.Flux

@Repository
class AnomaliesReadRepository(
    private val template: R2dbcEntityTemplate
) {

    fun findRecent(
        metricName: String?,
        algorithm: String?,
        approach: String?,
        limit: Int
    ): Flux<AnomalyRow> {
        var criteria = Criteria.empty()
        metricName?.let { criteria = criteria.and("metric_name").`is`(it) }
        algorithm?.let { criteria = criteria.and("algorithm").`is`(it) }
        approach?.let { criteria = criteria.and("approach").`is`(it) }

        val query = Query.query(criteria)
            .sort(Sort.by(Sort.Direction.DESC, "detected_at"))
            .limit(limit)

        return template.select(AnomalyRow::class.java).matching(query).all()
    }
}
