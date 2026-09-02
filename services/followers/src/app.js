const express = require("express");
const config = require("./config");
const followerRoutes = require("./routes/followerRoutes");
const { verifyConnection, ensureConstraints, driver } = require("./db");
const { logger, requestLogger, metricsMiddleware, metricsHandler, healthHandler } = require("./observability");
const { verifyToken } = require("./middleware/auth");

const app = express();
app.use(requestLogger);
app.use(metricsMiddleware);
app.use(express.json());
app.use(verifyToken);

app.get("/followers", (_req, res) => {
  res.json({ message: "Followers service radi" });
});

app.get("/health", healthHandler(driver));
app.get("/metrics", metricsHandler);

app.use("/followers", followerRoutes);

// eslint-disable-next-line no-unused-vars
app.use((err, req, res, _next) => {
  const status = err.status || 500;
  const log = req.log || logger;
  if (status >= 500) {
    log.error({ error: err.message }, "unhandled error");
  } else {
    log.warn({ error: err.message }, "request error");
  }
  res.status(status).json({ error: status >= 500 ? "Neocekivana greska" : err.message });
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
