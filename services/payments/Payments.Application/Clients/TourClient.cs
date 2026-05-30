using System.Net.Http.Json;

namespace Payments.Application.Clients;

public class TourClient
{
    private readonly HttpClient _httpClient;

    public TourClient(HttpClient httpClient)
    {
        _httpClient = httpClient;
    }

    public async Task<bool> IsTourPurchasableAsync(string tourId)
    {
       try
       {
           var response = await _httpClient.GetAsync($"/tours/{tourId}");

           if (!response.IsSuccessStatusCode)
               return false;

           var tour = await response.Content.ReadFromJsonAsync<TourResponse>();

           return tour != null && tour.Status == "published";
       }
       catch
       {
           return false;
       }

        if (!response.IsSuccessStatusCode)
            return false;

        var tour = await response.Content.ReadFromJsonAsync<TourResponse>();

        return tour != null && tour.Status == "published";
    }
}

public class TourResponse
{
    public string Id { get; set; } = string.Empty;
    public string Status { get; set; } = string.Empty;
}