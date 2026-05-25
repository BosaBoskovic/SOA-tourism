namespace Payments.Domain.Entities;

public class TourPurchaseToken
{
    public Guid Id { get; set; } = Guid.NewGuid();
    public string TouristId { get; set; } = string.Empty;
    public string TourId { get; set; } = string.Empty;
    public string TourName { get; set; } = string.Empty;
    public decimal Price { get; set; }
    public DateTime PurchasedAt { get; set; } = DateTime.UtcNow;
}