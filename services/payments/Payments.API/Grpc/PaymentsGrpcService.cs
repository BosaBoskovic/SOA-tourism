using Grpc.AspNetCore.Server;
using Grpc.Core;
using Microsoft.AspNetCore.Authorization;
using Payments.API.Auth;
using Payments.API.Grpc;
using Payments.Application.Services;

namespace Payments.API.Grpc;

[Authorize]
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
        var forbidden = CallerAuth.EnsureCallerIsTourist(context.GetHttpContext().User, request.TouristId);
        if (forbidden != null)
        {
            throw new RpcException(new Status(StatusCode.PermissionDenied, "You can only check out your own cart"));
        }

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
            throw new RpcException(
                new Status(StatusCode.FailedPrecondition, ex.Message));
        }
        catch (Exception ex)
        {
            throw new RpcException(
                new Status(StatusCode.Internal, ex.Message));
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
