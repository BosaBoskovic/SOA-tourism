namespace Payments.Domain.Entities;
using System.Text.Json.Serialization;

public class OrderItem
{
    public Guid Id { get; set; } = Guid.NewGuid();
    public Guid ShoppingCartId { get; set; }
    public string TourId { get; set; } = string.Empty;
    public string TourName { get; set; } = string.Empty;
    public decimal Price { get; set; }

    [JsonIgnore]
    public ShoppingCart ShoppingCart { get; set; } = null!;
}