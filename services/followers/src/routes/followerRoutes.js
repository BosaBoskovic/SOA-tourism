const express = require("express");
const controller = require("../controllers/followerController");

const router = express.Router();

router.post("/follow", controller.follow);
router.delete("/follow/:targetUsername", controller.unfollow);
router.get("/following/:username", controller.following);
router.get("/is-following", controller.isFollowing);
router.get("/visible-authors/:username", controller.visibleAuthors);
router.get("/recommendations/:username", controller.recommendations);

module.exports = router;
