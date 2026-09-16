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
        // Connection: close - stakeholders can run as multiple replicas
        // behind Docker's embedded DNS (docker-compose.yml no longer pins
        // its container_name). Node's fetch keeps HTTP connections alive
        // and reuses them by default, which would pin this call to
        // whichever replica answered first; closing after each call forces
        // a fresh connection (and DNS lookup) so these low-volume
        // notifications actually spread across replicas.
        headers: { "Content-Type": "application/json", "Connection": "close" },
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
