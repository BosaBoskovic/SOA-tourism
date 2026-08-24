FROM mcr.microsoft.com/dotnet/sdk:9.0 AS builder

WORKDIR /app

COPY Payments.Domain/Payments.Domain.csproj Payments.Domain/
COPY Payments.Infrastructure/Payments.Infrastructure.csproj Payments.Infrastructure/
COPY Payments.Application/Payments.Application.csproj Payments.Application/
COPY Payments.API/Payments.API.csproj Payments.API/

RUN dotnet restore Payments.API/Payments.API.csproj

COPY . .

RUN dotnet publish Payments.API/Payments.API.csproj -c Release -o /out

FROM mcr.microsoft.com/dotnet/aspnet:9.0

# curl for the container healthcheck below (base image has neither curl nor wget)
RUN apt-get update && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /out .

EXPOSE 8086

# Docker healthcheck oslanja se na /health endpoint (vidi Program.cs)
HEALTHCHECK --interval=15s --timeout=5s --start-period=20s --retries=5 \
    CMD curl -fsS http://localhost:8086/health || exit 1

ENTRYPOINT ["dotnet", "Payments.API.dll"]