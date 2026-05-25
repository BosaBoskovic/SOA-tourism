namespace Payments.Domain.Entities;

public class ShoppingCart
{
    public Guid Id { get; set; } = Guid.NewGuid();
    public string TouristId { get; set; } = string.Empty;
    public decimal TotalPrice { get; set; }
    public List<OrderItem> Items { get; set; } = new();

    public void RecalculateTotal()
    {
        TotalPrice = Items.Sum(i => i.Price);
    }
}