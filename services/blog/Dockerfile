FROM maven:3.9-eclipse-temurin-17 AS builder
WORKDIR /build
COPY pom.xml .
RUN mvn dependency:go-offline -q
COPY src ./src
RUN mvn package -DskipTests -q

FROM eclipse-temurin:17-jre-alpine
WORKDIR /app
COPY --from=builder /build/target/*.jar app.jar
EXPOSE 8082

# Docker healthcheck oslanja se na /health endpoint (actuator, vidi application.properties)
HEALTHCHECK --interval=15s --timeout=5s --start-period=30s --retries=5 \
    CMD wget -q -O- http://localhost:8082/health || exit 1

ENTRYPOINT ["java", "-jar", "app.jar"]