package com.flyte.analytics.anomaly

import java.time.Instant
import kotlin.math.sqrt

/**
 * EWMA control chart (Roberts, 1959): z_t = lambda * x_t + (1 - lambda) * z_{t-1}, z_0 = baseline.
 * Alarm when z leaves baseline -/+ width * sigma * sqrt(lambda / (2 - lambda)).
 *
 * Unlike CUSUM it has no memory of how long the deviation lasted, only of how large the recent
 * smoothed level is: a short burst of bad values decays away, a sustained shift pushes z past the limit.
 *
 * Asymptotic (constant) control limits are used on purpose. The exact limits are narrower for the
 * first observations after a reset, and on 0/1 observations (payment_success_rate) that turns
 * three failures in a row right after the runner's reset into a false positive.
 *
 * score = z itself, i.e. the smoothed metric value at the moment of detection.
 */
class EwmaDetector(
    private val baseline: Double,
    /** Standard deviation of a single observation in normal operation. */
    sigma: Double,
    private val lambda: Double,
    /** L, control limit width in standard deviations of z. */
    width: Double,
    private val direction: ChangeDirection = ChangeDirection.LOWER,
) : AnomalyDetector {

    init {
        require(lambda > 0.0 && lambda <= 1.0) { "EWMA lambda must be in (0, 1], got $lambda" }
        require(sigma > 0.0) { "EWMA sigma must be positive, got $sigma" }
        require(width > 0.0) { "EWMA width must be positive, got $width" }
    }

    private val limitHalfWidth = width * sigma * sqrt(lambda / (2.0 - lambda))

    private var ewma = baseline

    @Synchronized
    override fun observe(value: Double, timestamp: Instant): AnomalyResult {
        ewma = lambda * value + (1.0 - lambda) * ewma
        val isAnomaly = when (direction) {
            ChangeDirection.LOWER -> ewma < baseline - limitHalfWidth
            ChangeDirection.UPPER -> ewma > baseline + limitHalfWidth
        }
        return AnomalyResult(isAnomaly, timestamp, ewma)
    }

    @Synchronized
    override fun reset() {
        ewma = baseline
    }
}
