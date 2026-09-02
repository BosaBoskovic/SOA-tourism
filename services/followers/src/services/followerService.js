const neo4j = require("neo4j-driver");
const { driver } = require("../db");
const config = require("../config");

function toInteger(value) {
  if (neo4j.isInt(value)) {
    return value.toNumber();
  }
  return Number(value);
}

// accountExists checks stakeholders' :Account nodes (same physical Neo4j
// instance, different label) - the actual source of truth for "is this a
// real registered username", since followers' own :User nodes are just a
// lightweight social-graph projection created on first follow.
async function accountExists(session, username) {
  const result = await session.run(
    `MATCH (a:Account {username: $username}) RETURN a LIMIT 1`,
    { username }
  );
  return result.records.length > 0;
}

function httpError(message, status) {
  const err = new Error(message);
  err.status = status;
  return err;
}

async function followUser(followerUsername, targetUsername) {
  if (followerUsername === targetUsername) {
    throw httpError("User ne moze da zaprati samog sebe", 400);
  }

  const session = driver.session({ database: config.neo4jDatabase });
  try {
    const [followerExists, targetExists] = await Promise.all([
      accountExists(session, followerUsername),
      accountExists(session, targetUsername),
    ]);
    if (!followerExists || !targetExists) {
      throw httpError("Korisnik ne postoji", 404);
    }

    const result = await session.run(
      `
      MERGE (follower:User {username: $followerUsername})
      MERGE (target:User {username: $targetUsername})
      WITH follower, target, EXISTS { (follower)-[:FOLLOWS]->(target) } AS wasAlreadyFollowing
      MERGE (follower)-[r:FOLLOWS]->(target)
      ON CREATE SET r.createdAt = datetime()
      RETURN follower.username AS follower, target.username AS target, wasAlreadyFollowing
      `,
      { followerUsername, targetUsername }
    );

    const record = result.records[0];
    return {
      followerUsername: record.get("follower"),
      targetUsername: record.get("target"),
      alreadyFollowing: record.get("wasAlreadyFollowing"),
    };
  } finally {
    await session.close();
  }
}

async function unfollowUser(followerUsername, targetUsername) {
  const session = driver.session({ database: config.neo4jDatabase });
  try {
    await session.run(
      `
      MATCH (follower:User {username: $followerUsername})-[r:FOLLOWS]->(target:User {username: $targetUsername})
      DELETE r
      `,
      { followerUsername, targetUsername }
    );
  } finally {
    await session.close();
  }
}

async function isFollowing(followerUsername, targetUsername) {
  const session = driver.session({ database: config.neo4jDatabase });
  try {
    const result = await session.run(
      `
      MATCH (:User {username: $followerUsername})-[r:FOLLOWS]->(:User {username: $targetUsername})
      RETURN count(r) > 0 AS isFollowing
      `,
      { followerUsername, targetUsername }
    );

    if (result.records.length === 0) {
      return false;
    }

    return Boolean(result.records[0].get("isFollowing"));
  } finally {
    await session.close();
  }
}

async function getFollowing(userUsername) {
  const session = driver.session({ database: config.neo4jDatabase });
  try {
    const result = await session.run(
      `
      MATCH (:User {username: $userUsername})-[:FOLLOWS]->(followed:User)
      RETURN followed.username AS username
      ORDER BY username
      `,
      { userUsername }
    );

    return result.records.map((record) => record.get("username"));
  } finally {
    await session.close();
  }
}

// getFollowers is the inverse of getFollowing: who follows userUsername,
// rather than who userUsername follows.
async function getFollowers(userUsername) {
  const session = driver.session({ database: config.neo4jDatabase });
  try {
    const result = await session.run(
      `
      MATCH (follower:User)-[:FOLLOWS]->(:User {username: $userUsername})
      RETURN follower.username AS username
      ORDER BY username
      `,
      { userUsername }
    );

    return result.records.map((record) => record.get("username"));
  } finally {
    await session.close();
  }
}

async function getVisibleAuthors(userUsername) {
  const following = await getFollowing(userUsername);
  const all = [userUsername, ...following];
  return Array.from(new Set(all));
}

async function getRecommendations(userUsername, limit = 10) {
  const session = driver.session({ database: config.neo4jDatabase });
  try {
    const result = await session.run(
      `
      MATCH (:User {username: $userUsername})-[:FOLLOWS]->(:User)-[:FOLLOWS]->(candidate:User)
      WHERE candidate.username <> $userUsername
      AND NOT EXISTS {
        MATCH (:User {username: $userUsername})-[:FOLLOWS]->(candidate)
      }
      RETURN candidate.username AS username, count(*) AS score
      ORDER BY score DESC, username ASC
      LIMIT $limit
      `,
      { userUsername, limit: neo4j.int(limit) }
    );

    return result.records.map((record) => ({
      username: record.get("username"),
      score: toInteger(record.get("score")),
    }));
  } finally {
    await session.close();
  }
}

module.exports = {
  followUser,
  unfollowUser,
  isFollowing,
  getFollowing,
  getFollowers,
  getVisibleAuthors,
  getRecommendations,
};
