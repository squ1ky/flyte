package com.flyte.analytics.stream

import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertEquals

class WindowMathTest {

    @Test
    fun `truncates to the start of a 60 second window`() {
        val instant = Instant.parse("2026-01-01T00:01:47Z")
        assertEquals(Instant.parse("2026-01-01T00:01:00Z"), WindowMath.truncateToWindow(instant, 60))
    }

    @Test
    fun `instant already at window boundary is unchanged`() {
        val instant = Instant.parse("2026-01-01T00:02:00Z")
        assertEquals(instant, WindowMath.truncateToWindow(instant, 60))
    }

    @Test
    fun `works with non-default window sizes`() {
        val instant = Instant.parse("2026-01-01T00:00:37Z")
        assertEquals(Instant.parse("2026-01-01T00:00:30Z"), WindowMath.truncateToWindow(instant, 30))
    }
}
