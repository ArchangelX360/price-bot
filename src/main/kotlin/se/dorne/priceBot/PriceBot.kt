package se.dorne.priceBot

import io.ktor.application.*
import io.ktor.metrics.micrometer.*
import io.ktor.response.*
import io.ktor.routing.*
import io.ktor.server.engine.*
import io.ktor.server.netty.*
import io.micrometer.prometheus.PrometheusConfig
import io.micrometer.prometheus.PrometheusMeterRegistry
import kotlinx.coroutines.asCoroutineDispatcher
import kotlinx.coroutines.runBlocking
import se.dorne.priceBot.utils.FetcherScheduler
import se.dorne.priceBot.websites.Megekko
import java.util.concurrent.Executors
import kotlin.time.ExperimentalTime
import kotlin.time.hours

object BotDispatcher {
    val ScheduledFetcher = Executors.newFixedThreadPool(5).asCoroutineDispatcher()
}

@OptIn(ExperimentalTime::class)
fun main(): Unit = runBlocking {
    val appMicrometerRegistry = PrometheusMeterRegistry(PrometheusConfig.DEFAULT)
    FetcherScheduler(
        fetcher = Megekko(
            appMicrometerRegistry,
        ),
        registry = appMicrometerRegistry,
        maxRetry = 2,
        period = 1.hours, // Megekko does not allow more than 1 hours interval for scrapers
    ).startIn(this)

    embeddedServer(Netty, port = 8080, host = "127.0.0.1") {
        install(MicrometerMetrics) {
            registry = appMicrometerRegistry
        }
        routing {
            get("/metrics") {
                call.respond(appMicrometerRegistry.scrape())
            }
        }
    }.start(wait = true)
}
