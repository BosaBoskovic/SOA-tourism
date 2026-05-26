using Microsoft.AspNetCore.Mvc;
using Payments.API.DTOs;
using Payments.Application.Services;

namespace Payments.API.Controllers;

[ApiController]
[Route("shopping-cart")]
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
        var cart = await _cartService.GetCartAsync(touristId);
        return Ok(cart);
    }

    // POST /shopping-cart/{touristId}/items
    [HttpPost("{touristId}/items")]
    public async Task<IActionResult> AddItem(string touristId, [FromBody] AddItemRequest req)
    {
        try
        {
            var cart = await _cartService.AddItemAsync(touristId, req.TourId, req.TourName, req.Price);
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