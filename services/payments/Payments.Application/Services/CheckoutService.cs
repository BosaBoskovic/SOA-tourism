using Payments.Application.Clients;
using Payments.Domain.Entities;
using Payments.Infrastructure.Repositories;
using Payments.Application.Events;
using Payments.Infrastructure.Messaging;

namespace Payments.Application.Services;

public class CheckoutService
{
    private readonly ShoppingCartRepository _cartRepo;
    private readonly TourPurchaseTokenRepository _tokenRepo;
    private readonly TourClient _tourClient;
    private readonly RabbitMqPublisher _publisher;

    public CheckoutService(
        ShoppingCartRepository cartRepo,
        TourPurchaseTokenRepository tokenRepo,
        TourClient tourClient,
        RabbitMqPublisher publisher)
    {
        _cartRepo = cartRepo;
        _tokenRepo = tokenRepo;
        _tourClient = tourClient;
        _publisher = publisher;
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


            var completedEvent = new PurchaseCompletedEvent
            {
                SagaId = Guid.NewGuid().ToString(),
                TouristId = touristId,
                Items = tokens.Select(t => new PurchasedTourItem
                {
                    TourId = t.TourId,
                    TourName = t.TourName,
                    Price = (double)t.Price
                }).ToList()
            };

            _publisher.Publish("purchase-completed", completedEvent);

            return tokens;
        }
        catch
        {
            await _tokenRepo.DeleteRangeAsync(tokens);
            throw;
        }
    }

    public async Task<string> StartCheckoutSagaAsync(string touristId)
    {
        var cart = await _cartRepo.GetByTouristIdAsync(touristId);

        if (cart == null || cart.Items.Count == 0)
            throw new InvalidOperationException("Korpa je prazna.");

        var sagaId = Guid.NewGuid().ToString();

        var eventMessage = new CheckoutRequestedEvent
        {
            SagaId = sagaId,
            TouristId = touristId,
            Items = cart.Items.Select(i => new CheckoutTourItem
            {
                TourId = i.TourId,
                TourName = i.TourName,
                Price = (double)i.Price
            }).ToList()
        };

        _publisher.Publish("checkout-requested", eventMessage);

        return sagaId;
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