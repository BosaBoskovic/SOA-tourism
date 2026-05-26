using Microsoft.AspNetCore.Mvc;
using Payments.Application.Services;

namespace Payments.API.Controllers;

[ApiController]
[Route("checkout")]
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
        var tokens = await _checkoutService.GetPurchasedToursAsync(touristId);
        return Ok(tokens);
    }

    // GET /checkout/{touristId}/has-purchased/{tourId}
    [HttpGet("{touristId}/has-purchased/{tourId}")]
    public async Task<IActionResult> HasPurchased(string touristId, string tourId)
    {
        var result = await _checkoutService.HasPurchasedAsync(touristId, tourId);
        return Ok(new { hasPurchased = result });
    }
}