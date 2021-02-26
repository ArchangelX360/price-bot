package se.dorne.priceBot

import io.ktor.application.*
import io.ktor.metrics.micrometer.*
import io.ktor.response.*
import io.ktor.routing.*
import io.ktor.server.engine.*
import io.ktor.server.netty.*
import io.micrometer.core.instrument.DistributionSummary
import io.micrometer.core.instrument.ImmutableTag
import io.micrometer.prometheus.PrometheusConfig
import io.micrometer.prometheus.PrometheusMeterRegistry
import kotlinx.coroutines.asCoroutineDispatcher
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.Serializable
import kotlinx.serialization.descriptors.PrimitiveKind
import org.hildan.chrome.devtools.domains.runtime.evaluateJs
import org.hildan.chrome.devtools.protocol.ChromeDPClient
import org.hildan.chrome.devtools.targets.attachToNewPageAndAwaitPageLoad
import org.hildan.chrome.devtools.targets.use
import se.dorne.priceBot.utils.Fetcher
import se.dorne.priceBot.utils.FetcherScheduler
import java.util.concurrent.Executors
import java.util.concurrent.atomic.AtomicInteger
import kotlin.time.ExperimentalTime
import kotlin.time.seconds

object BotDispatcher {
    val ScheduledFetcher = Executors.newFixedThreadPool(5).asCoroutineDispatcher()
}

data class ProductListing(
    val url: String,
    val category: Category,
) {
    enum class Category(val prometheusLabel: String) {
        GPU("gpu"),
        CPU("cpu");
    }
}

@Serializable
data class Product(val name: String, val price: Float)

class Megekko(
    private val registry: PrometheusMeterRegistry,
) : Fetcher<Unit> {
    private val productLists = listOf(
        ProductListing(
            url = "https://www.megekko.nl/Computer/Componenten/Videokaarten/Nvidia-Videokaarten?f=f_429-170345,169343,169344,169345_vrrd-0_s-populair_pp-250_p-1_d-list_cf-",
            category = ProductListing.Category.GPU,
        ),
        ProductListing(
            url = "https://www.megekko.nl/Computer/Componenten/Processoren/Socket-AM4-Processoren?f=f_58-145914,151027_vrrd-0_s-populair_pp-50_p-1_d-list_cf-",
            category = ProductListing.Category.CPU,
        ),
    )

    override suspend fun fetch() {
        productLists.forEach { listing ->
            ChromeDPClient("http://127.0.0.1:9222").webSocket().use { browserSession ->
                browserSession
                    .attachToNewPageAndAwaitPageLoad(
                        url = listing.url,
                    )
                    .use { pageSession ->
                        val getProductPricesQuery = """
                            Array
                                .from(document.querySelectorAll(".container"))
                                .map(n => {
                                    const name = n.querySelector("h2.title")?.innerText
                                    const price = parseFloat(n.querySelector(".pricecontainer")?.innerText)
                                    if (name && price && !isNaN(price)) {
                                        return {name, price}
                                    } else {
                                        return null
                                    }
                                })
                                .filter(listing => listing !== null)

                        """.trimIndent()
                        val products = pageSession.runtime.evaluateJs<List<Product>>(getProductPricesQuery)
                        products?.forEach { product ->
                            registry.gauge(
                                "bot_prices", mutableListOf(
                                    ImmutableTag("website", name),
                                    ImmutableTag("category", listing.category.prometheusLabel),
                                    ImmutableTag("name", product.name),
                                ), AtomicInteger(),
                            )!!.set(product.price.toInt())
                        }
                    }
            }
        }
    }
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
        period = 10.seconds,
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
