const config = require("../config");

// Best-effort: a notification failing (stakeholders down/slow) should never
// break the follow/unfollow request that triggered it.
async function notify(username, type, message, relatedUsername) {
  try {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 3000);
    try {
      await fetch(`${config.stakeholdersUrl}/stakeholders/notifications/internal`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, type, message, relatedUsername }),
        signal: controller.signal,
      });
    } finally {
      clearTimeout(timeout);
    }
  } catch (err) {
    console.error(JSON.stringify({ level: "warn", msg: "notification failed", error: String(err) }));
  }
}

async function notifyNewFollower(followerUsername, targetUsername) {
  await notify(
    targetUsername,
    "follow",
    `${followerUsername} je počeo/la da te prati.`,
    followerUsername
  );
}

module.exports = { notifyNewFollower };
