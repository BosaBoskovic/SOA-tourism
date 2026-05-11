const express = require("express");
const config = require("./config");
const followerRoutes = require("./routes/followerRoutes");
const { verifyConnection, ensureConstraints, driver } = require("./db");

const app = express();
app.use(express.json());

app.get("/followers", (_req, res) => {
  res.json({ message: "Followers service radi" });
});

app.use("/followers", followerRoutes);

app.use((err, _req, res, _next) => {
  console.error(err);
  res.status(500).json({ error: "Neocekivana greska" });
});

async function bootstrap() {
  await verifyConnection();
  await ensureConstraints();

  app.listen(config.port, () => {
    console.log(`Followers service running on port ${config.port}`);
  });
}

bootstrap().catch(async (error) => {
  console.error("Neuspelo pokretanje followers servisa", error);
  await driver.close();
  process.exit(1);
});

process.on("SIGTERM", async () => {
  await driver.close();
  process.exit(0);
});
