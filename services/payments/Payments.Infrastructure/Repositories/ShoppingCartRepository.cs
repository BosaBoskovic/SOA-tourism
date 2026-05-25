using Microsoft.EntityFrameworkCore;
using Payments.Domain.Entities;
using Payments.Infrastructure.Data;

namespace Payments.Infrastructure.Repositories;

public class ShoppingCartRepository
{
    private readonly PaymentsDbContext _db;

    public ShoppingCartRepository(PaymentsDbContext db)
    {
        _db = db;
    }

    public async Task<ShoppingCart?> GetByTouristIdAsync(string touristId)
    {
        return await _db.ShoppingCarts
            .Include(c => c.Items)
            .FirstOrDefaultAsync(c => c.TouristId == touristId);
    }

    public async Task<ShoppingCart> GetOrCreateAsync(string touristId)
    {
        var cart = await GetByTouristIdAsync(touristId);
        if (cart != null) return cart;

        cart = new ShoppingCart { TouristId = touristId };
        _db.ShoppingCarts.Add(cart);
        await _db.SaveChangesAsync();
        return cart;
    }
    public async Task AddItemAsync(OrderItem item)
    {
        await _db.OrderItems.AddAsync(item);
    }

    public async Task SaveAsync(ShoppingCart cart)
    {
        await _db.SaveChangesAsync();
    }

    public async Task RemoveItemAsync(ShoppingCart cart, Guid itemId)
    {
        var item = cart.Items.FirstOrDefault(i => i.Id == itemId);
        if (item == null) return;

        _db.OrderItems.Remove(item);
        cart.Items.Remove(item);
        cart.RecalculateTotal();
        //_db.ShoppingCarts.Update(cart);
        await _db.SaveChangesAsync();
    }

    public async Task ClearAsync(ShoppingCart cart)
    {
        _db.OrderItems.RemoveRange(cart.Items);
        cart.Items.Clear();
        cart.TotalPrice = 0;
        //_db.ShoppingCarts.Update(cart);
        await _db.SaveChangesAsync();
    }
}