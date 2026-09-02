using System.Net;
using System.Net.Http.Json;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging.Abstractions;
using Moq;
using Payments.Application.Clients;
using Payments.Application.Services;
using Payments.Domain.Entities;
using Payments.Infrastructure.Data;
using Payments.Infrastructure.Messaging;
using Payments.Infrastructure.Repositories;
using Xunit;

namespace Payments.Tests;

// CheckoutAsync is exercised against a real (InMemory) EF Core context and
// the real repositories, not mocks of them - the behavior under test (one
// DB transaction, idempotent re-checkout) is how those pieces interact
// together, not something any single method does alone. TourClient is real
// too, pointed at a fake HttpMessageHandler instead of a real tours
// service. RabbitMqPublisher is real but points at an address nothing is
// listening on - CheckoutAsync already treats a failed publish as
// non-fatal (logged, not thrown), so a fast-failing connection attempt is
// exactly the case a "don't let messaging block a paid purchase" test
// should cover anyway.
public class CheckoutServiceTests : IDisposable
{
    private readonly PaymentsDbContext _db;

    public CheckoutServiceTests()
    {
        var options = new DbContextOptionsBuilder<PaymentsDbContext>()
            .UseInMemoryDatabase(Guid.NewGuid().ToString())
            .Options;
        _db = new PaymentsDbContext(options);
    }

    public void Dispose() => _db.Dispose();

    private static TourClient FakeTourClient(Func<HttpRequestMessage, HttpResponseMessage> respond)
    {
        var httpClient = new HttpClient(new StubHttpMessageHandler(respond))
        {
            BaseAddress = new Uri("http://tours.test")
        };
        return new TourClient(httpClient);
    }

    private static RabbitMqPublisher UnreachableRabbitMqPublisher()
    {
        var config = new Mock<IConfiguration>();
        config.Setup(c => c["RABBITMQ_HOST"]).Returns("127.0.0.1");
        return new RabbitMqPublisher(config.Object);
    }

    private CheckoutService BuildService(Func<HttpRequestMessage, HttpResponseMessage> tourResponder)
    {
        return new CheckoutService(
            _db,
            new ShoppingCartRepository(_db),
            new TourPurchaseTokenRepository(_db),
            FakeTourClient(tourResponder),
            UnreachableRabbitMqPublisher(),
            NullLogger<CheckoutService>.Instance);
    }

    private static HttpResponseMessage PublishedTourResponse(string tourId, string name, decimal price)
    {
        return new HttpResponseMessage(HttpStatusCode.OK)
        {
            Content = JsonContent.Create(new { id = tourId, status = "published", name, price })
        };
    }

    [Fact]
    public async Task CheckoutAsync_PurchasesAllItemsAndClearsTheCart()
    {
        var cart = new ShoppingCart { TouristId = "ana" };
        cart.Items.Add(new OrderItem { ShoppingCartId = cart.Id, TourId = "tour-1", TourName = "Kalemegdan", Price = 10m });
        cart.Items.Add(new OrderItem { ShoppingCartId = cart.Id, TourId = "tour-2", TourName = "Ada Ciganlija", Price = 20m });
        _db.ShoppingCarts.Add(cart);
        await _db.SaveChangesAsync();

        var service = BuildService(_ => PublishedTourResponse("ignored", "ignored", 0));

        var tokens = await service.CheckoutAsync("ana");

        Assert.Equal(2, tokens.Count);
        Assert.Contains(tokens, t => t.TourId == "tour-1");
        Assert.Contains(tokens, t => t.TourId == "tour-2");

        var remainingCart = await _db.ShoppingCarts.Include(c => c.Items).FirstAsync(c => c.TouristId == "ana");
        Assert.Empty(remainingCart.Items);

        var persisted = await _db.TourPurchaseTokens.Where(t => t.TouristId == "ana").ToListAsync();
        Assert.Equal(2, persisted.Count);
    }

    [Fact]
    public async Task CheckoutAsync_IsIdempotent_DoesNotRepurchaseAnAlreadyOwnedTour()
    {
        var cart = new ShoppingCart { TouristId = "ana" };
        cart.Items.Add(new OrderItem { ShoppingCartId = cart.Id, TourId = "tour-1", TourName = "Kalemegdan", Price = 10m });
        _db.ShoppingCarts.Add(cart);
        _db.TourPurchaseTokens.Add(new TourPurchaseToken { TouristId = "ana", TourId = "tour-1", TourName = "Kalemegdan", Price = 10m });
        await _db.SaveChangesAsync();

        var service = BuildService(_ => PublishedTourResponse("tour-1", "Kalemegdan", 10m));

        var tokens = await service.CheckoutAsync("ana");

        Assert.Single(tokens);
        var tokensForTour = await _db.TourPurchaseTokens
            .Where(t => t.TouristId == "ana" && t.TourId == "tour-1")
            .ToListAsync();
        Assert.Single(tokensForTour); // re-checking out never creates a duplicate token
    }

    [Fact]
    public async Task CheckoutAsync_ThrowsOnEmptyCart()
    {
        var service = BuildService(_ => PublishedTourResponse("x", "x", 0));

        await Assert.ThrowsAsync<InvalidOperationException>(() => service.CheckoutAsync("nobody-with-a-cart"));
    }

    [Fact]
    public async Task CheckoutAsync_RejectsATourThatIsNoLongerPurchasable_AndLeavesTheCartIntact()
    {
        var cart = new ShoppingCart { TouristId = "ana" };
        cart.Items.Add(new OrderItem { ShoppingCartId = cart.Id, TourId = "tour-1", TourName = "Kalemegdan", Price = 10m });
        _db.ShoppingCarts.Add(cart);
        await _db.SaveChangesAsync();

        var service = BuildService(_ => new HttpResponseMessage(HttpStatusCode.NotFound));

        await Assert.ThrowsAsync<InvalidOperationException>(() => service.CheckoutAsync("ana"));

        // Rejected before the transaction ever opens - nothing should have changed.
        var remainingCart = await _db.ShoppingCarts.Include(c => c.Items).FirstAsync(c => c.TouristId == "ana");
        Assert.Single(remainingCart.Items);
        Assert.Empty(await _db.TourPurchaseTokens.ToListAsync());
    }

    private class StubHttpMessageHandler : HttpMessageHandler
    {
        private readonly Func<HttpRequestMessage, HttpResponseMessage> _respond;

        public StubHttpMessageHandler(Func<HttpRequestMessage, HttpResponseMessage> respond)
        {
            _respond = respond;
        }

        protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
        {
            return Task.FromResult(_respond(request));
        }
    }
}
