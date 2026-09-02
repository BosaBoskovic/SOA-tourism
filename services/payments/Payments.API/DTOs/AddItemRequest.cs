namespace Payments.API.DTOs;

// TourName/Price used to be client-supplied here and were trusted outright
// (a caller could "buy" anything for any price they typed). Both now always
// come from the tour itself (see ShoppingCartService.AddItemAsync) - only
// the id is meaningful from the client's side.
public class AddItemRequest
{
    public string TourId { get; set; } = string.Empty;
}
