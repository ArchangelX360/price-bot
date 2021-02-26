FROM openjdk:11.0.2-jre

EXPOSE 8080:8080

RUN mkdir /app
COPY bin/ /app/bin/
COPY lib/ /app/lib/

WORKDIR /app/bin
CMD ["./price-bot"]
