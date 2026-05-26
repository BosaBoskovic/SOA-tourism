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

WORKDIR /app
COPY --from=builder /out .

EXPOSE 8086
ENTRYPOINT ["dotnet", "Payments.API.dll"]