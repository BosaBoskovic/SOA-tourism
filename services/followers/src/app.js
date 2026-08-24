const express = require("express");
const config = require("./config");
const followerRoutes = require("./routes/followerRoutes");
const { verifyConnection, ensureConstraints, driver } = require("./db");
const { logger, requestLogger, metricsMiddleware, metricsHandler, healthHandler } = require("./observability");

const app = express();
app.use(requestLogger);
app.use(metricsMiddleware);
app.use(express.json());

app.get("/followers", (_req, res) => {
  res.json({ message: "Followers service radi" });
});

app.get("/health", healthHandler(driver));
app.get("/metrics", metricsHandler);

app.use("/followers", followerRoutes);

app.use((err, req, res, _next) => {
  req.log ? req.log.error({ error: err.message }, "unhandled error") : logger.error({ error: err.message }, "unhandled error");
  res.status(500).json({ error: "Neocekivana greska" });
});

async function bootstrap() {
  await verifyConnection();
  await ensureConstraints();

  app.listen(config.port, () => {
    logger.info({ port: config.port }, "followers service starting");
  });
}

bootstrap().catch(async (error) => {
  logger.error({ error: error.message }, "neuspelo pokretanje followers servisa");
  await driver.close();
  process.exit(1);
});

process.on("SIGTERM", async () => {
  await driver.close();
  process.exit(0);
});
