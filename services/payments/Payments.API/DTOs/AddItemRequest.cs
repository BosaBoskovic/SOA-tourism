namespace Payments.API.DTOs;

public class AddItemRequest
{
    public string TourId { get; set; } = string.Empty;
    public string TourName { get; set; } = string.Empty;
    public decimal Price { get; set; }
}