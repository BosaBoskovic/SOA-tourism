const followerService = require("../services/followerService");

function getActorUsername(req) {
  return (req.header("X-Username") || "").trim();
}

function requireActor(req, res) {
  const actorUsername = getActorUsername(req);
  if (!actorUsername) {
    res.status(401).json({ error: "Nedostaje X-Username header" });
    return null;
  }
  return actorUsername;
}

async function follow(req, res) {
  const actorUsername = requireActor(req, res);
  if (!actorUsername) {
    return;
  }

  const targetUsername = (req.body.targetUsername || "").trim();
  if (!targetUsername) {
    res.status(400).json({ error: "targetUsername je obavezan" });
    return;
  }

  try {
    const relation = await followerService.followUser(actorUsername, targetUsername);
    res.status(201).json({ message: "Uspesno pracenje", relation });
  } catch (error) {
    res.status(400).json({ error: error.message });
  }
}

async function unfollow(req, res) {
  const actorUsername = requireActor(req, res);
  if (!actorUsername) {
    return;
  }

  const targetUsername = (req.params.targetUsername || "").trim();
  if (!targetUsername) {
    res.status(400).json({ error: "targetUsername je obavezan" });
    return;
  }

  await followerService.unfollowUser(actorUsername, targetUsername);
  res.status(204).send();
}

async function following(req, res) {
  const username = (req.params.username || "").trim();
  if (!username) {
    res.status(400).json({ error: "username je obavezan" });
    return;
  }

  const users = await followerService.getFollowing(username);
  res.json({ username, following: users });
}

async function isFollowing(req, res) {
  const followerUsername = (req.query.followerUsername || "").trim();
  const targetUsername = (req.query.targetUsername || "").trim();

  if (!followerUsername || !targetUsername) {
    res.status(400).json({ error: "followerUsername i targetUsername su obavezni" });
    return;
  }

  const follows = await followerService.isFollowing(followerUsername, targetUsername);
  res.json({ followerUsername, targetUsername, isFollowing: follows });
}

async function visibleAuthors(req, res) {
  const username = (req.params.username || "").trim();
  if (!username) {
    res.status(400).json({ error: "username je obavezan" });
    return;
  }

  const authors = await followerService.getVisibleAuthors(username);
  res.json({ username, authors });
}

async function recommendations(req, res) {
  const username = (req.params.username || "").trim();
  const limit = Number(req.query.limit) || 10;

  if (!username) {
    res.status(400).json({ error: "username je obavezan" });
    return;
  }

  const suggested = await followerService.getRecommendations(username, limit);
  res.json({ username, recommendations: suggested });
}

module.exports = {
  follow,
  unfollow,
  following,
  isFollowing,
  visibleAuthors,
  recommendations,
};
