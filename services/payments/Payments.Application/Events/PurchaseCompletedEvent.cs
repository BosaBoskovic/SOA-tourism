namespace Payments.Application.Events;

public class PurchaseCompletedEvent
{
    public string SagaId { get; set; } = string.Empty;
    public string TouristId { get; set; } = string.Empty;
    public List<PurchasedTourItem> Items { get; set; } = new();
}

public class PurchasedTourItem
{
    public string TourId { get; set; } = string.Empty;
    public string TourName { get; set; } = string.Empty;
    public double Price { get; set; }
}
