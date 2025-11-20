# 1. Build
FROM maven:3.9.11-eclipse-temurin-25 AS builder

WORKDIR /app

COPY pom.xml ./pom.xml
COPY flight-service ./flight-service

RUN mvn -B -q -pl flight-service -am clean package -DskipTests

# 2. Runtime
FROM eclipse-temurin:25-jre AS runtime

WORKDIR /app

COPY --from=builder /app/flight-service/target/flight-service-0.0.1-SNAPSHOT.jar app.jar

EXPOSE 8081

ENTRYPOINT ["java", "-jar", "app.jar"]
