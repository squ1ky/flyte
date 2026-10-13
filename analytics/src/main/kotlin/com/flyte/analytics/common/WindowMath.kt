package com.flyte.analytics.common

import java.time.Instant

object WindowMath {
    fun truncateToWindow(instant: Instant, windowSizeSeconds: Int): Instant {
        val epochSecond = instant.epochSecond
        val truncated = epochSecond - (epochSecond % windowSizeSeconds)
        return Instant.ofEpochSecond(truncated)
    }
}