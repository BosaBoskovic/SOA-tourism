namespace Payments.Application.Events;

public class CheckoutRequestedEvent
{
    public string SagaId { get; set; } = string.Empty;
    public string TouristId { get; set; } = string.Empty;
    public List<CheckoutTourItem> Items { get; set; } = new();
}

public class CheckoutTourItem
{
    public string TourId { get; set; } = string.Empty;
    public string TourName { get; set; } = string.Empty;
    public double Price { get; set; }
}

public class CheckoutApprovedEvent
{
    public string SagaId { get; set; } = string.Empty;
    public string TouristId { get; set; } = string.Empty;
}

public class CheckoutRejectedEvent
{
    public string SagaId { get; set; } = string.Empty;
    public string Reason { get; set; } = string.Empty;
}

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