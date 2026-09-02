const express = require("express");
const controller = require("../controllers/followerController");
const asyncHandler = require("../utils/asyncHandler");

const router = express.Router();

router.post("/follow", asyncHandler(controller.follow));
router.delete("/follow/:targetUsername", asyncHandler(controller.unfollow));
router.get("/following/:username", asyncHandler(controller.following));
router.get("/is-following", asyncHandler(controller.isFollowing));
router.get("/visible-authors/:username", asyncHandler(controller.visibleAuthors));
router.get("/recommendations/:username", asyncHandler(controller.recommendations));

module.exports = router;
