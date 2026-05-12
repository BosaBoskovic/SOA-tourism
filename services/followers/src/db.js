const neo4j = require("neo4j-driver");
const config = require("./config");

const driver = neo4j.driver(
  config.neo4jUri,
  neo4j.auth.basic(config.neo4jUser, config.neo4jPassword)
);

async function verifyConnection() {
  await driver.verifyConnectivity();
}

async function ensureConstraints() {
  const session = driver.session({ database: config.neo4jDatabase });
  try {
    await session.run(
      "CREATE CONSTRAINT user_username_unique IF NOT EXISTS FOR (u:User) REQUIRE u.username IS UNIQUE"
    );
  } finally {
    await session.close();
  }
}

module.exports = {
  driver,
  verifyConnection,
  ensureConstraints,
};
