using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using Payments.API.Auth;
using Payments.Application.Services;

namespace Payments.API.Controllers;

[ApiController]
[Route("checkout")]
[Authorize]
public class CheckoutController : ControllerBase
{
    private readonly CheckoutService _checkoutService;

    public CheckoutController(CheckoutService checkoutService)
    {
        _checkoutService = checkoutService;
    }

    // POST /checkout/{touristId}
    [HttpPost("{touristId}")]
    public async Task<IActionResult> Checkout(string touristId)
    {
        if (CallerAuth.EnsureCallerIsTourist(User, touristId) is { } forbidden) return forbidden;

        try
        {
            var tokens = await _checkoutService.CheckoutAsync(touristId);
            return Ok(tokens);
        }
        catch (InvalidOperationException ex)
        {
            return BadRequest(new { error = ex.Message });
        }
    }

    // GET /checkout/{touristId}/purchases
    [HttpGet("{touristId}/purchases")]
    public async Task<IActionResult> GetPurchases(string touristId)
    {
        if (CallerAuth.EnsureCallerIsTourist(User, touristId) is { } forbidden) return forbidden;

        var tokens = await _checkoutService.GetPurchasedToursAsync(touristId);
        return Ok(tokens);
    }

    // POST /checkout/analytics - guide-facing purchase count/revenue per
    // tour. Any authenticated caller can ask (the response only aggregates
    // numbers, no per-tourist identity), matching HasPurchased's low
    // sensitivity rather than needing an ownership check like the tourist-scoped endpoints above.
    [HttpPost("analytics")]
    public async Task<IActionResult> GetAnalytics([FromBody] List<string> tourIds)
    {
        if (tourIds == null || tourIds.Count == 0)
        {
            return Ok(new List<TourAnalytics>());
        }
        var analytics = await _checkoutService.GetAnalyticsAsync(tourIds);
        return Ok(analytics);
    }

    // GET /checkout/{touristId}/has-purchased/{tourId}
    // Called service-to-service by tours (to gate review creation) with no
    // bearer token, so this one stays open rather than requiring auth -
    // it only discloses a yes/no purchase flag, not cart/purchase contents.
    [HttpGet("{touristId}/has-purchased/{tourId}")]
    [AllowAnonymous]
    public async Task<IActionResult> HasPurchased(string touristId, string tourId)
    {
        var result = await _checkoutService.HasPurchasedAsync(touristId, tourId);
        return Ok(new { hasPurchased = result });
    }
}
