using System.Security.Claims;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;

namespace Payments.API.Auth;

/// <summary>
/// Every action here is scoped to a single tourist's cart/purchases; the
/// route's touristId must match the verified caller (or the caller must be
/// an admin) instead of being trusted outright.
/// </summary>
public static class CallerAuth
{
    public static IActionResult? EnsureCallerIsTourist(ClaimsPrincipal user, string touristId)
    {
        var callerUsername = user.FindFirstValue("sub");
        var callerRole = user.FindFirstValue("role");

        if (callerRole == "admin")
        {
            return null;
        }
        if (callerUsername == touristId)
        {
            return null;
        }
        return new ObjectResult(new { error = "You can only manage your own cart/purchases" })
        {
            StatusCode = StatusCodes.Status403Forbidden
        };
    }
}
