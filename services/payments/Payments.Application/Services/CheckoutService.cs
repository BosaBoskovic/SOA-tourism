using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;
using Payments.Application.Clients;
using Payments.Domain.Entities;
using Payments.Infrastructure.Data;
using Payments.Infrastructure.Repositories;
using Payments.Application.Events;
using Payments.Infrastructure.Messaging;

namespace Payments.Application.Services;

public class CheckoutService
{
    private readonly PaymentsDbContext _db;
    private readonly ShoppingCartRepository _cartRepo;
    private readonly TourPurchaseTokenRepository _tokenRepo;
    private readonly TourClient _tourClient;
    private readonly RabbitMqPublisher _publisher;
    private readonly ILogger<CheckoutService> _logger;

    public CheckoutService(
        PaymentsDbContext db,
        ShoppingCartRepository cartRepo,
        TourPurchaseTokenRepository tokenRepo,
        TourClient tourClient,
        RabbitMqPublisher publisher,
        ILogger<CheckoutService> logger)
    {
        _db = db;
        _cartRepo = cartRepo;
        _tokenRepo = tokenRepo;
        _tourClient = tourClient;
        _publisher = publisher;
        _logger = logger;
    }

    // Safe to call more than once for the same cart (e.g. a retried
    // request): tours the caller already holds a token for are skipped
    // rather than re-purchased, backed by a unique (TouristId, TourId)
    // index so a race between two concurrent calls can't double-purchase either.
    public async Task<List<TourPurchaseToken>> CheckoutAsync(string touristId)
    {
        var cart = await _cartRepo.GetByTouristIdAsync(touristId);

        if (cart == null || cart.Items.Count == 0)
            throw new InvalidOperationException("Korpa je prazna.");

        var alreadyPurchased = (await _tokenRepo.GetByTouristIdAsync(touristId))
            .Where(t => cart.Items.Any(i => i.TourId == t.TourId))
            .ToList();
        var alreadyPurchasedTourIds = alreadyPurchased.Select(t => t.TourId).ToHashSet();

        var itemsToPurchase = cart.Items.Where(i => !alreadyPurchasedTourIds.Contains(i.TourId)).ToList();

        foreach (var item in itemsToPurchase)
        {
            var canBuy = await _tourClient.IsTourPurchasableAsync(item.TourId);

            if (!canBuy)
                throw new InvalidOperationException($"Tura '{item.TourName}' nije dostupna za kupovinu.");
        }

        var newTokens = itemsToPurchase.Select(item => new TourPurchaseToken
        {
            TouristId = touristId,
            TourId = item.TourId,
            TourName = item.TourName,
            Price = item.Price,
            PurchasedAt = DateTime.UtcNow
        }).ToList();

        // Token creation and clearing the cart happen in one transaction:
        // either both happen, or neither does - no more "tokens rolled back
        // but the cart is already gone" on a mid-checkout failure.
        await using (var transaction = await _db.Database.BeginTransactionAsync())
        {
            if (newTokens.Count > 0)
            {
                await _tokenRepo.AddRangeAsync(newTokens);
            }
            await _cartRepo.ClearAsync(cart);
            await transaction.CommitAsync();
        }

        // The purchase is already durably committed at this point - a
        // failure to publish the notification is logged, not treated as a
        // reason to undo a real purchase.
        if (newTokens.Count > 0)
        {
            try
            {
                var completedEvent = new PurchaseCompletedEvent
                {
                    SagaId = Guid.NewGuid().ToString(),
                    TouristId = touristId,
                    Items = newTokens.Select(t => new PurchasedTourItem
                    {
                        TourId = t.TourId,
                        TourName = t.TourName,
                        Price = (double)t.Price
                    }).ToList()
                };
                _publisher.Publish("purchase-completed", completedEvent);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to publish purchase-completed for tourist {TouristId}", touristId);
            }
        }

        return alreadyPurchased.Concat(newTokens).ToList();
    }

    public async Task<List<TourPurchaseToken>> GetPurchasedToursAsync(string touristId)
    {
        return await _tokenRepo.GetByTouristIdAsync(touristId);
    }

    public async Task<bool> HasPurchasedAsync(string touristId, string tourId)
    {
        return await _tokenRepo.HasPurchasedAsync(touristId, tourId);
    }

    // Guide-facing analytics: purchase count + revenue per tour.
    public async Task<List<TourAnalytics>> GetAnalyticsAsync(IEnumerable<string> tourIds)
    {
        var analytics = await _tokenRepo.GetAnalyticsAsync(tourIds);
        return analytics.Select(kvp => new TourAnalytics
        {
            TourId = kvp.Key,
            PurchaseCount = kvp.Value.Count,
            Revenue = kvp.Value.Revenue,
        }).ToList();
    }
}

public class TourAnalytics
{
    public string TourId { get; set; } = string.Empty;
    public int PurchaseCount { get; set; }
    public decimal Revenue { get; set; }
}
