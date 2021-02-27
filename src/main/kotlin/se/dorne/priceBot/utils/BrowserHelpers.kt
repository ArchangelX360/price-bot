package se.dorne.priceBot.utils

import org.hildan.chrome.devtools.domains.network.SetUserAgentOverrideRequest
import org.hildan.chrome.devtools.targets.ChromePageSession
import org.hildan.chrome.devtools.targets.use

suspend inline fun <R> ChromePageSession.useWithUserAgent(userAgent: String, block: (ChromePageSession) -> R): R = use {
    it.network.setUserAgentOverride(SetUserAgentOverrideRequest(userAgent = userAgent))
    block(this)
}
