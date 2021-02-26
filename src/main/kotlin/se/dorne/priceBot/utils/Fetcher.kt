package se.dorne.priceBot.utils

import io.micrometer.core.instrument.ImmutableTag
import io.micrometer.core.instrument.MeterRegistry
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.filterNotNull
import kotlinx.coroutines.flow.first
import org.slf4j.LoggerFactory
import se.dorne.priceBot.BotDispatcher
import java.time.Instant
import java.util.concurrent.atomic.AtomicLong
import kotlin.time.Duration
import kotlin.time.ExperimentalTime

interface Fetcher<T : Any> { // : Any to prevent T to be nullable
    val name: String get() = this::class.simpleName ?: "anonymous"

    suspend fun fetch(): T
}

@OptIn(ExperimentalTime::class, ExperimentalCoroutinesApi::class)
class FetcherScheduler<T : Any>(
    // : Any to prevent T to be nullable
    private val fetcher: Fetcher<T>,
    private val period: Duration,
    registry: MeterRegistry,
    private val maxRetry: Int = 5,
) {
    private val lastSuccessfulRefresh: AtomicLong = registry.gauge(
        "bot_fetcher_last_successful_refresh",
        mutableListOf(ImmutableTag("fetcher_name", fetcher.name)),
        AtomicLong(),
    )!!

    private var state: MutableStateFlow<T?> = MutableStateFlow(null)

    fun startIn(scope: CoroutineScope) {
        scope.launch(BotDispatcher.ScheduledFetcher + CoroutineName(fetcher.name)) {
            LOG.info("[${fetcher.name}] [period: $period] Starting fetcher scheduler")
            while (isActive) {
                val v = withTimeoutOrNull(period) {
                    updateState()
                }
                if (v == null) {
                    LOG.warn("[${fetcher.name}] timed out fetching, will retry")
                    continue // skip delay
                }
                delay(period)
            }
        }
    }

    private suspend fun updateState() {
        repeat(maxRetry) {
            try {
                LOG.info("[${fetcher.name}] Fetching...")
                state.value = fetcher.fetch()
                lastSuccessfulRefresh.set(Instant.now().epochSecond)
                LOG.info("[${fetcher.name}] successfully fetched, will run again in $period")
                return
            } catch (e: TimeoutCancellationException) {
                throw e
            } catch (e: Exception) {
                val remainingRetries = maxRetry - 1 - it
                if (remainingRetries > 0) {
                    LOG.warn("[${fetcher.name}] failed to fetch (${e::class.simpleName}), retrying $remainingRetries more time(s)")
                } else {
                    LOG.error("[${fetcher.name}] failed to fetch with:", e)
                }
            }
        }
    }

    suspend fun getValue(): T = state.filterNotNull().first()

    companion object {
        private val LOG = LoggerFactory.getLogger(FetcherScheduler::class.java)
    }
}
