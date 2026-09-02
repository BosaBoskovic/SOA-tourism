const jwt = require("jsonwebtoken");

const secret = process.env.JWT_SECRET;
if (!secret) {
  throw new Error("JWT_SECRET is not set; followers cannot verify tokens");
}

// verifyToken checks the same HS256 access tokens stakeholders issues and
// attaches the caller's identity to the request, instead of trusting the
// X-Username header the gateway sets (which itself never checks the JWT
// signature). It never rejects by itself - requireActor does that, so it
// stays a plain verify-only middleware.
function verifyToken(req, _res, next) {
  const header = (req.header("Authorization") || "").trim();
  const match = /^Bearer\s+(.+)$/i.exec(header);
  if (match) {
    try {
      const decoded = jwt.verify(match[1], secret);
      req.verifiedUsername = decoded.sub;
      req.verifiedRole = decoded.role;
    } catch (err) {
      // invalid/expired token: leave unset, requireActor rejects with 401
    }
  }
  next();
}

// requireActor returns the verified caller's username, or writes 401 and
// returns null.
function requireActor(req, res) {
  if (!req.verifiedUsername) {
    res.status(401).json({ error: "Autentifikacija je obavezna" });
    return null;
  }
  return req.verifiedUsername;
}

module.exports = { verifyToken, requireActor };
