using Payments.Application.Clients;
using Payments.Domain.Entities;
using Payments.Infrastructure.Repositories;

namespace Payments.Application.Services;

public class ShoppingCartService
{
    private readonly ShoppingCartRepository _cartRepo;
    private readonly TourClient _tourClient;

    public ShoppingCartService(ShoppingCartRepository cartRepo, TourClient tourClient)
    {
        _cartRepo = cartRepo;
        _tourClient = tourClient;
    }

    public async Task<ShoppingCart> GetCartAsync(string touristId)
    {
        return await _cartRepo.GetOrCreateAsync(touristId);
    }

    // tourId is the only thing about the item that's actually trusted from
    // the client - name and price always come from the tour itself, never
    // from the request body, so a cart (and the checkout that follows it)
    // can't be built with a forged price.
    public async Task<ShoppingCart> AddItemAsync(string touristId, string tourId)
    {
        var tour = await _tourClient.GetPurchasableTourAsync(tourId)
            ?? throw new InvalidOperationException("Tura nije dostupna za kupovinu.");

        var cart = await _cartRepo.GetOrCreateAsync(touristId);

        // Proveri da li je tura već u korpi
        if (cart.Items.Any(i => i.TourId == tourId))
            throw new InvalidOperationException("Ova tura je već u korpi.");

        var item = new OrderItem
        {
            ShoppingCartId = cart.Id,
            TourId = tourId,
            TourName = tour.Name,
            Price = tour.Price
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
