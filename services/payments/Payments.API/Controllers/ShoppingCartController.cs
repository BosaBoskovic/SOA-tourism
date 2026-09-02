using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using Payments.API.Auth;
using Payments.API.DTOs;
using Payments.Application.Services;

namespace Payments.API.Controllers;

[ApiController]
[Route("shopping-cart")]
[Authorize]
public class ShoppingCartController : ControllerBase
{
    private readonly ShoppingCartService _cartService;

    public ShoppingCartController(ShoppingCartService cartService)
    {
        _cartService = cartService;
    }

    // GET /shopping-cart/{touristId}
    [HttpGet("{touristId}")]
    public async Task<IActionResult> GetCart(string touristId)
    {
        if (CallerAuth.EnsureCallerIsTourist(User, touristId) is { } forbidden) return forbidden;

        var cart = await _cartService.GetCartAsync(touristId);
        return Ok(cart);
    }

    // POST /shopping-cart/{touristId}/items
    [HttpPost("{touristId}/items")]
    public async Task<IActionResult> AddItem(string touristId, [FromBody] AddItemRequest req)
    {
        if (CallerAuth.EnsureCallerIsTourist(User, touristId) is { } forbidden) return forbidden;

        try
        {
            var cart = await _cartService.AddItemAsync(touristId, req.TourId);
            return Ok(cart);
        }
        catch (InvalidOperationException ex)
        {
            return BadRequest(new { error = ex.Message });
        }
    }

    // DELETE /shopping-cart/{touristId}/items/{itemId}
    [HttpDelete("{touristId}/items/{itemId}")]
    public async Task<IActionResult> RemoveItem(string touristId, Guid itemId)
    {
        if (CallerAuth.EnsureCallerIsTourist(User, touristId) is { } forbidden) return forbidden;

        try
        {
            var cart = await _cartService.RemoveItemAsync(touristId, itemId);
            return Ok(cart);
        }
        catch (InvalidOperationException ex)
        {
            return BadRequest(new { error = ex.Message });
        }
    }
}
