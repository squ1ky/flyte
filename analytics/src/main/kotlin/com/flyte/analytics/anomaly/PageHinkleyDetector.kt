package com.flyte.analytics.anomaly

import java.time.Instant

/**
 * Page-Hinkley test (Page, 1954; Hinkley, 1971) with the reference level learned from the data.
 *
 * PH_t = max(0, PH_{t-1} + d_t), d_t = (mean_t - x_t) - delta for a drop (LOWER),
 * d_t = (x_t - mean_t) - delta for a rise (UPPER); alarm when PH_t > threshold.
 * That is the CUSUM recursion with the configured baseline replaced by the running mean of the
 * observations themselves, so the pair isolates one question: known vs learned normal level.
 * Page-Hinkley needs no baseline in the config, but it slowly adapts to the fault it is detecting.
 *
 * The mean is a fading average: weight_t = alpha * weight_{t-1} + 1, mean_t moves towards x_t by
 * 1 / weight_t. alpha = 1 is the classic cumulative mean; alpha < 1 limits the memory to about
 * 1 / (1 - alpha) observations, so fault periods of previous runs do not drag the reference down
 * for the rest of the experiment.
 *
 * [reset] clears only the PH statistic: the learned mean is knowledge about normal behaviour, and
 * the runner resets right before injecting a fault, so a cleared mean would be learned from the fault.
 * No alarm is raised until [minObservations] values have been seen since startup.
 *
 * score = PH_t.
 */
class PageHinkleyDetector(
    private val threshold: Double,
    private val delta: Double,
    private val alpha: Double,
    private val minObservations: Int,
    private val direction: ChangeDirection = ChangeDirection.LOWER,
) : AnomalyDetector {

    init {
        require(alpha > 0.0 && alpha <= 1.0) { "Page-Hinkley alpha must be in (0, 1], got $alpha" }
        require(minObservations >= 1) { "Page-Hinkley minObservations must be >= 1, got $minObservations" }
    }

    private var mean = 0.0
    private var weight = 0.0
    private var observations = 0L
    private var statistic = 0.0

    @Synchronized
    override fun observe(value: Double, timestamp: Instant): AnomalyResult {
        weight = alpha * weight + 1.0
        mean += (value - mean) / weight
        observations++

        if (observations < minObservations) return AnomalyResult(false, timestamp, 0.0)

        val deviation = when (direction) {
            ChangeDirection.LOWER -> mean - value
            ChangeDirection.UPPER -> value - mean
        }
        statistic = maxOf(0.0, statistic + deviation - delta)
        return AnomalyResult(statistic > threshold, timestamp, statistic)
    }

    @Synchronized
    override fun reset() {
        statistic = 0.0
    }
}
