using Grpc.Core;
using Payments.API.Grpc;
using Payments.Application.Services;

namespace Payments.API.Grpc;

public class PaymentsGrpcService : PaymentsService.PaymentsServiceBase
{
    private readonly CheckoutService _checkoutService;

    public PaymentsGrpcService(CheckoutService checkoutService)
    {
        _checkoutService = checkoutService;
    }

    public override async Task<CheckoutResponse> Checkout(
        CheckoutRequest request, ServerCallContext context)
    {
        try
        {
            var tokens = await _checkoutService.CheckoutAsync(request.TouristId);

            var response = new CheckoutResponse();
            foreach (var token in tokens)
            {
                response.Tokens.Add(new PurchaseToken
                {
                    Id = token.Id.ToString(),
                    TouristId = token.TouristId,
                    TourId = token.TourId,
                    TourName = token.TourName,
                    Price = (double)token.Price,
                    PurchasedAt = token.PurchasedAt.ToString("o")
                });
            }
            return response;
        }
        catch (InvalidOperationException ex)
        {
            throw new RpcException(new Status(StatusCode.InvalidArgument, ex.Message));
        }
    }

    public override async Task<HasPurchasedResponse> HasPurchased(
        HasPurchasedRequest request, ServerCallContext context)
    {
        var result = await _checkoutService.HasPurchasedAsync(
            request.TouristId, request.TourId);

        return new HasPurchasedResponse { HasPurchased = result };
    }
}