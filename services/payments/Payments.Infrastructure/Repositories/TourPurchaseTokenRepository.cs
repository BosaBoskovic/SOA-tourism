using Microsoft.EntityFrameworkCore;
using Payments.Domain.Entities;
using Payments.Infrastructure.Data;

namespace Payments.Infrastructure.Repositories;

public class TourPurchaseTokenRepository
{
    private readonly PaymentsDbContext _db;

    public TourPurchaseTokenRepository(PaymentsDbContext db)
    {
        _db = db;
    }

    public async Task AddRangeAsync(IEnumerable<TourPurchaseToken> tokens)
    {
        await _db.TourPurchaseTokens.AddRangeAsync(tokens);
        await _db.SaveChangesAsync();
    }

    public async Task<List<TourPurchaseToken>> GetByTouristIdAsync(string touristId)
    {
        return await _db.TourPurchaseTokens
            .Where(t => t.TouristId == touristId)
            .OrderByDescending(t => t.PurchasedAt)
            .ToListAsync();
    }

    public async Task<bool> HasPurchasedAsync(string touristId, string tourId)
    {
        return await _db.TourPurchaseTokens
            .AnyAsync(t => t.TouristId == touristId && t.TourId == tourId);
    }

    public async Task DeleteRangeAsync(IEnumerable<TourPurchaseToken> tokens)
    {
        _db.TourPurchaseTokens.RemoveRange(tokens);
        await _db.SaveChangesAsync();
    }
}