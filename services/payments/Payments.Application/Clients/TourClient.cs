using System.Net.Http.Json;
using System.Text.Json;

namespace Payments.Application.Clients;

public class TourClient
{
    // tours' JSON fields are camelCase ("status", "price", ...); the default
    // JsonSerializerOptions.Default used by ReadFromJsonAsync is case-sensitive
    // and would silently leave every property at its default value otherwise.
    private static readonly JsonSerializerOptions JsonOptions = new() { PropertyNameCaseInsensitive = true };

    private readonly HttpClient _httpClient;

    public TourClient(HttpClient httpClient)
    {
        _httpClient = httpClient;
    }

    public async Task<bool> IsTourPurchasableAsync(string tourId)
    {
        return await GetPurchasableTourAsync(tourId) != null;
    }

    // The only place a tour's name/price should come from - never the
    // client-supplied AddItemRequest, which can't be trusted for either.
    // Returns null if the tour doesn't exist or isn't published.
    public async Task<TourResponse?> GetPurchasableTourAsync(string tourId)
    {
        try
        {
            var response = await _httpClient.GetAsync($"/tours/{tourId}");

            if (!response.IsSuccessStatusCode)
                return null;

            var tour = await response.Content.ReadFromJsonAsync<TourResponse>(JsonOptions);

            return tour != null && tour.Status == "published" ? tour : null;
        }
        catch
        {
            return null;
        }
    }
}

public class TourResponse
{
    public string Id { get; set; } = string.Empty;
    public string Status { get; set; } = string.Empty;
    public string Name { get; set; } = string.Empty;
    public decimal Price { get; set; }
}
