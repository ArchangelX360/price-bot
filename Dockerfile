FROM docker.pkg.github.com/archangelx360/price-bot/base

ADD price-bot .

EXPOSE 8091

ENTRYPOINT ["sh", "-c", "./price-bot"]
