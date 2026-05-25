using Payments.Domain.Entities;
using Payments.Infrastructure.Repositories;

namespace Payments.Application.Services;

public class CheckoutService
{
    private readonly ShoppingCartRepository _cartRepo;
    private readonly TourPurchaseTokenRepository _tokenRepo;

    public CheckoutService(ShoppingCartRepository cartRepo, TourPurchaseTokenRepository tokenRepo)
    {
        _cartRepo = cartRepo;
        _tokenRepo = tokenRepo;
    }

    public async Task<List<TourPurchaseToken>> CheckoutAsync(string touristId)
    {
        var cart = await _cartRepo.GetByTouristIdAsync(touristId);

        if (cart == null || cart.Items.Count == 0)
            throw new InvalidOperationException("Korpa je prazna.");

        // Generiši token za svaku stavku
        var tokens = cart.Items.Select(item => new TourPurchaseToken
        {
            TouristId = touristId,
            TourId = item.TourId,
            TourName = item.TourName,
            Price = item.Price,
            PurchasedAt = DateTime.UtcNow
        }).ToList();

        await _tokenRepo.AddRangeAsync(tokens);
        await _cartRepo.ClearAsync(cart);

        return tokens;
    }

    public async Task<List<TourPurchaseToken>> GetPurchasedToursAsync(string touristId)
    {
        return await _tokenRepo.GetByTouristIdAsync(touristId);
    }

    public async Task<bool> HasPurchasedAsync(string touristId, string tourId)
    {
        return await _tokenRepo.HasPurchasedAsync(touristId, tourId);
    }
}