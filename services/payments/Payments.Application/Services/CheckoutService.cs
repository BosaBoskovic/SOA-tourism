using Payments.Application.Clients;
using Payments.Domain.Entities;
using Payments.Infrastructure.Repositories;

namespace Payments.Application.Services;

public class CheckoutService
{
    private readonly ShoppingCartRepository _cartRepo;
    private readonly TourPurchaseTokenRepository _tokenRepo;
    private readonly TourClient _tourClient;

    public CheckoutService(
        ShoppingCartRepository cartRepo,
        TourPurchaseTokenRepository tokenRepo,
        TourClient tourClient)
    {
        _cartRepo = cartRepo;
        _tokenRepo = tokenRepo;
        _tourClient = tourClient;
    }

    public async Task<List<TourPurchaseToken>> CheckoutAsync(string touristId)
    {
        var cart = await _cartRepo.GetByTouristIdAsync(touristId);

        if (cart == null || cart.Items.Count == 0)
            throw new InvalidOperationException("Korpa je prazna.");

        foreach (var item in cart.Items)
        {
            var canBuy = await _tourClient.IsTourPurchasableAsync(item.TourId);

            if (!canBuy)
                throw new InvalidOperationException($"Tura '{item.TourName}' nije dostupna za kupovinu.");
        }

        var tokens = cart.Items.Select(item => new TourPurchaseToken
        {
            TouristId = touristId,
            TourId = item.TourId,
            TourName = item.TourName,
            Price = item.Price,
            PurchasedAt = DateTime.UtcNow
        }).ToList();

        try
        {
            await _tokenRepo.AddRangeAsync(tokens);
            await _cartRepo.ClearAsync(cart);

            return tokens;
        }
        catch
        {
            await _tokenRepo.DeleteRangeAsync(tokens);
            throw;
        }
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