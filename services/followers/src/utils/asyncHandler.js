// Wraps an async route handler so a rejected promise reaches Express's
// error middleware via next(err) instead of crashing the process with an
// unhandled rejection (Node terminates on those by default).
function asyncHandler(fn) {
  return (req, res, next) => {
    Promise.resolve(fn(req, res, next)).catch(next);
  };
}

module.exports = asyncHandler;
