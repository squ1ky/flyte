package com.flyte.analytics.anomaly

import java.time.Instant
import kotlin.math.sqrt
import kotlin.random.Random
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

/**
 * Monte Carlo check of the detector parameters in application.yaml (keep them in sync).
 *
 * Every metric is fed one 0/1 observation per event, so the model is exact: Bernoulli(p) in normal
 * operation. Model of one experiment run: 2 min warmup + 3 min baseline at 0.5 bookings/s = 60 + 90
 * observations, then a total outage (all 0s, or all 1s for abandonment). The runner resets detectors
 * before the baseline and before the injection, so does the simulation.
 *
 * p comes from the live baseline: payment_success_rate 0.8 is
 * FAKE_BANK_SUCCESS_CHANCE; conversion_rate 0.76 = 0.8 x (1 - 5% user cancellations), measured 0.760
 * over 603 outcomes; abandonment_rate_expired 0 (no expirations at all in normal operation).
 */
class DetectorCalibrationTest {

    private val now = Instant.parse("2026-01-01T00:00:00Z")

    private data class Calibration(val falsePositiveRate: Double, val maxObservationsToDetect: Int)

    private fun lowerDetectors(p: Double, cusumDrift: Double): Map<String, () -> AnomalyDetector> = mapOf(
        AnomalyAlgorithms.CUSUM to { CusumDetector(baseline = p, threshold = 8.0, drift = cusumDrift) },
        AnomalyAlgorithms.EWMA to {
            EwmaDetector(baseline = p, sigma = sqrt(p * (1 - p)), lambda = 0.05, width = 4.0)
        },
        AnomalyAlgorithms.PAGE_HINKLEY to {
            PageHinkleyDetector(threshold = 8.0, delta = 0.05, alpha = 0.99, minObservations = 30)
        },
    )

    private fun simulate(create: () -> AnomalyDetector, p: Double, faultValue: Double, runs: Int = 2000): Calibration {
        val random = Random(42)
        fun normal() = if (random.nextDouble() < p) 1.0 else 0.0

        var falsePositives = 0
        var maxDelay = 0
        repeat(runs) {
            val detector = create()
            repeat(60) { detector.observe(normal(), now) }

            detector.reset()
            if ((1..90).count { detector.observe(normal(), now).isAnomaly } > 0) falsePositives++

            detector.reset()
            val delay = generateSequence(1) { it + 1 }.first { detector.observe(faultValue, now).isAnomaly }
            maxDelay = maxOf(maxDelay, delay)
        }
        return Calibration(falsePositives.toDouble() / runs, maxDelay)
    }

    private fun assertCalibrated(metric: String, p: Double, cusumDrift: Double) {
        lowerDetectors(p, cusumDrift).forEach { (algorithm, create) ->
            val calibration = simulate(create, p, faultValue = 0.0)
            println("$metric $algorithm: $calibration")
            // Comparable false positive rates (all < 1%), otherwise comparing speed is unfair.
            assertTrue(calibration.falsePositiveRate < 0.01, "$metric $algorithm: $calibration")
            // 16 outcomes at 0.5/s = 32 s of detection delay on top of the pipeline latency.
            assertTrue(calibration.maxObservationsToDetect <= 16, "$metric $algorithm: $calibration")
        }
    }

    @Test
    fun `payment_success_rate parameters keep false positives rare and detection fast`() {
        assertCalibrated(AnomalyMetricNames.PAYMENT_SUCCESS_RATE, p = 0.8, cusumDrift = 0.05)
    }

    @Test
    fun `conversion_rate parameters keep false positives rare and detection fast`() {
        // drift 0.05 gives 1.4% false positives at p = 0.76, several times more than EWMA/PH.
        assertCalibrated(AnomalyMetricNames.CONVERSION_RATE, p = 0.76, cusumDrift = 0.07)
    }

    @Test
    fun `abandonment_rate_expired detectors all need the same number of expirations in a row`() {
        // Normal level is exactly 0, so a false positive is impossible for any threshold; what matters
        // is that the three algorithms are equally sensitive: 4 consecutive expirations.
        val detectors = mapOf(
            AnomalyAlgorithms.CUSUM to CusumDetector(0.0, threshold = 3.0, drift = 0.01, ChangeDirection.UPPER),
            AnomalyAlgorithms.EWMA to EwmaDetector(0.0, sigma = 0.44, lambda = 0.1, width = 3.0, ChangeDirection.UPPER),
            AnomalyAlgorithms.PAGE_HINKLEY to PageHinkleyDetector(3.0, delta = 0.01, alpha = 0.99, minObservations = 30, ChangeDirection.UPPER),
        )
        detectors.forEach { (algorithm, detector) ->
            repeat(150) { assertTrue(!detector.observe(0.0, now).isAnomaly, "$algorithm fired on normal traffic") }
            detector.reset()
            val delay = generateSequence(1) { it + 1 }.first { detector.observe(1.0, now).isAnomaly }
            assertEquals(4, delay, algorithm)
        }
    }
}
