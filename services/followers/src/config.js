const config = {
  port: process.env.PORT || 8084,
  neo4jUri: process.env.NEO4J_URI || "bolt://localhost:7687",
  neo4jUser: process.env.NEO4J_USER || "neo4j",
  neo4jPassword: process.env.NEO4J_PASSWORD || "password",
  neo4jDatabase: process.env.NEO4J_DATABASE || "neo4j",
};

module.exports = config;
