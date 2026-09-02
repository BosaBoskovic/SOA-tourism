const followerService = require("../services/followerService");
const { requireActor } = require("../middleware/auth");
const { notifyNewFollower } = require("../services/notifier");

const MIN_RECOMMENDATIONS_LIMIT = 1;
const MAX_RECOMMENDATIONS_LIMIT = 100;
const DEFAULT_RECOMMENDATIONS_LIMIT = 10;

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

  const relation = await followerService.followUser(actorUsername, targetUsername);
  if (!relation.alreadyFollowing) {
    notifyNewFollower(actorUsername, targetUsername); // fire-and-forget
  }
  res.status(relation.alreadyFollowing ? 200 : 201).json({
    message: relation.alreadyFollowing ? "Vec pratite ovog korisnika" : "Uspesno pracenje",
    relation,
  });
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
  if (!requireActor(req, res)) {
    return;
  }

  const username = (req.params.username || "").trim();
  if (!username) {
    res.status(400).json({ error: "username je obavezan" });
    return;
  }

  const users = await followerService.getFollowing(username);
  res.json({ username, following: users });
}

async function followers(req, res) {
  if (!requireActor(req, res)) {
    return;
  }

  const username = (req.params.username || "").trim();
  if (!username) {
    res.status(400).json({ error: "username je obavezan" });
    return;
  }

  const users = await followerService.getFollowers(username);
  res.json({ username, followers: users });
}

// isFollowing and visibleAuthors are also called service-to-service by blog
// (which has no bearer token to forward for these specific lookups), so
// they intentionally stay open rather than requiring a verified actor -
// network lockdown (only reachable via the gateway/internal network) is the
// mitigation for these two, not per-call auth.
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
  if (!requireActor(req, res)) {
    return;
  }

  const username = (req.params.username || "").trim();
  if (!username) {
    res.status(400).json({ error: "username je obavezan" });
    return;
  }

  let limit = Number.parseInt(req.query.limit, 10);
  if (!Number.isFinite(limit)) {
    limit = DEFAULT_RECOMMENDATIONS_LIMIT;
  }
  limit = Math.min(Math.max(limit, MIN_RECOMMENDATIONS_LIMIT), MAX_RECOMMENDATIONS_LIMIT);

  const suggested = await followerService.getRecommendations(username, limit);
  res.json({ username, recommendations: suggested });
}

module.exports = {
  follow,
  unfollow,
  following,
  followers,
  isFollowing,
  visibleAuthors,
  recommendations,
};
