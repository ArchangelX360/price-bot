package se.dorne.priceBot.websites

import io.micrometer.core.instrument.ImmutableTag
import io.micrometer.prometheus.PrometheusMeterRegistry
import kotlinx.serialization.Serializable
import org.hildan.chrome.devtools.domains.runtime.evaluateJs
import org.hildan.chrome.devtools.protocol.ChromeDPClient
import org.hildan.chrome.devtools.targets.attachToNewPageAndAwaitPageLoad
import org.hildan.chrome.devtools.targets.use
import org.slf4j.LoggerFactory
import se.dorne.priceBot.utils.useWithUserAgent
import se.dorne.priceBot.utils.Fetcher
import java.util.concurrent.atomic.AtomicInteger

class Megekko(
    private val registry: PrometheusMeterRegistry,
) : Fetcher<Unit> {
    private val fakeUserAgent =
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.192 Safari/537.36"

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
            ChromeDPClient().webSocket().use { browserSession ->
                browserSession
                    .attachToNewPageAndAwaitPageLoad(
                        url = listing.url,
                    )
                    .useWithUserAgent(fakeUserAgent) { pageSession ->
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
                        LOG.info("{}", products)
                        products?.forEach { product ->
                            registry.gauge(
                                "bot_prices",
                                mutableListOf(
                                    ImmutableTag("website", name),
                                    ImmutableTag("category", listing.category.prometheusLabel),
                                    ImmutableTag("name", product.name),
                                ),
                                AtomicInteger(),
                            )!!.set(product.price.toInt())
                        }
                    }
            }
        }
    }

    companion object {
        private val LOG = LoggerFactory.getLogger(Megekko::class.java)
    }
}

private data class ProductListing(
    val url: String,
    val category: Category,
) {
    enum class Category(val prometheusLabel: String) {
        GPU("gpu"),
        CPU("cpu");
    }
}

@Serializable
private data class Product(val name: String, val price: Float)
