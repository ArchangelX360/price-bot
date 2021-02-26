import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
    val kotlinVersion = "1.4.30"
    kotlin("jvm") version kotlinVersion
    kotlin("plugin.serialization") version kotlinVersion
    id("com.palantir.docker") version "0.26.0"
    id("com.palantir.git-version") version "0.12.3"
    application
}

group = "se.dorne"
version = "1.0-SNAPSHOT"

repositories {
    jcenter()
    mavenCentral()
    maven { url = uri("https://dl.bintray.com/kotlin/kotlinx") }
    maven { url = uri("https://dl.bintray.com/kotlin/ktor") }
}

dependencies {
    testImplementation(kotlin("test-junit"))
    val ktorVersion = "1.4.0"
    implementation("io.ktor:ktor-server-netty:$ktorVersion")
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.1.0")

    implementation("io.ktor:ktor-metrics-micrometer:$ktorVersion")
    implementation("io.micrometer:micrometer-registry-prometheus:1.6.4")

    implementation("org.hildan.chrome:chrome-devtools-kotlin:1.0.0")

    implementation("org.slf4j:slf4j-simple:1.7.30")
}

tasks.test {
    useJUnit()
}

tasks.withType<KotlinCompile> {
    kotlinOptions.jvmTarget = "1.8"
    kotlinOptions.freeCompilerArgs = listOf("-Xopt-in=kotlin.RequiresOptIn")
}

application {
    mainClassName = "se.dorne.priceBot.PriceBotKt"
}

docker {
    val registry = "docker.pkg.github.com"
    val repository = "archangelx360/price-bot"
    val imageFullName = "$registry/$repository/${project.name}"

    val versionDetails: groovy.lang.Closure<com.palantir.gradle.gitversion.VersionDetails> by extra
    val ciCommitSha: String? = System.getenv("GITHUB_SHA") ?: versionDetails().gitHashFull

    dependsOn(tasks.test.get())
    dependsOn(tasks.installDist.get())

    name = imageFullName

    tag("latestSha", "$imageFullName:$ciCommitSha")

    files(tasks.installDist.get().outputs)

    pull(true)
}
