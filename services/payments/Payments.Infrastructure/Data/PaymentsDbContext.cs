using Microsoft.EntityFrameworkCore;
using Payments.Domain.Entities;

namespace Payments.Infrastructure.Data;

public class PaymentsDbContext : DbContext
{
    public PaymentsDbContext(DbContextOptions<PaymentsDbContext> options) : base(options) { }

    public DbSet<ShoppingCart> ShoppingCarts => Set<ShoppingCart>();
    public DbSet<OrderItem> OrderItems => Set<OrderItem>();
    public DbSet<TourPurchaseToken> TourPurchaseTokens => Set<TourPurchaseToken>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<ShoppingCart>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.Property(e => e.TouristId).IsRequired();
            entity.Property(e => e.TotalPrice).HasColumnType("decimal(18,2)");
            entity.HasMany(e => e.Items)
                  .WithOne(e => e.ShoppingCart)
                  .HasForeignKey(e => e.ShoppingCartId)
                  .OnDelete(DeleteBehavior.Cascade);
        });

        modelBuilder.Entity<OrderItem>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.Property(e => e.TourId).IsRequired();
            entity.Property(e => e.TourName).IsRequired();
            entity.Property(e => e.Price).HasColumnType("decimal(18,2)");
        });

        modelBuilder.Entity<TourPurchaseToken>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.Property(e => e.TouristId).IsRequired();
            entity.Property(e => e.TourId).IsRequired();
            entity.Property(e => e.Price).HasColumnType("decimal(18,2)");
        });
    }
}