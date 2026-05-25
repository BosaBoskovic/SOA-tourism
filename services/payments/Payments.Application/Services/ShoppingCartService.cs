using Payments.Domain.Entities;
using Payments.Infrastructure.Repositories;

namespace Payments.Application.Services;

public class ShoppingCartService
{
    private readonly ShoppingCartRepository _cartRepo;

    public ShoppingCartService(ShoppingCartRepository cartRepo)
    {
        _cartRepo = cartRepo;
    }

    public async Task<ShoppingCart> GetCartAsync(string touristId)
    {
        return await _cartRepo.GetOrCreateAsync(touristId);
    }

    public async Task<ShoppingCart> AddItemAsync(string touristId, string tourId, string tourName, decimal price)
    {
        var cart = await _cartRepo.GetOrCreateAsync(touristId);

        // Proveri da li je tura već u korpi
        if (cart.Items.Any(i => i.TourId == tourId))
            throw new InvalidOperationException("Ova tura je već u korpi.");

        var item = new OrderItem
        {
            ShoppingCartId = cart.Id,
            TourId = tourId,
            TourName = tourName,
            Price = price
        };

        cart.Items.Add(item);
        cart.RecalculateTotal();

        await _cartRepo.AddItemAsync(item);

        await _cartRepo.SaveAsync(cart);
        return cart;
    }

    public async Task<ShoppingCart> RemoveItemAsync(string touristId, Guid itemId)
    {
        var cart = await _cartRepo.GetByTouristIdAsync(touristId)
            ?? throw new InvalidOperationException("Korpa nije pronađena.");

        await _cartRepo.RemoveItemAsync(cart, itemId);
        return await _cartRepo.GetByTouristIdAsync(touristId)!;
    }

    public async Task ClearCartAsync(string touristId)
    {
        var cart = await _cartRepo.GetByTouristIdAsync(touristId);
        if (cart == null) return;
        await _cartRepo.ClearAsync(cart);
    }
}