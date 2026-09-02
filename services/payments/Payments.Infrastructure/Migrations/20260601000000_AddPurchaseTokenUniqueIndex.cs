using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Payments.Infrastructure.Migrations
{
    /// <inheritdoc />
    public partial class AddPurchaseTokenUniqueIndex : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            // A tourist can only ever hold one purchase token per tour - this is
            // what makes checkout safe to retry (see CheckoutService.CheckoutAsync).
            migrationBuilder.CreateIndex(
                name: "IX_TourPurchaseTokens_TouristId_TourId",
                table: "TourPurchaseTokens",
                columns: new[] { "TouristId", "TourId" },
                unique: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropIndex(
                name: "IX_TourPurchaseTokens_TouristId_TourId",
                table: "TourPurchaseTokens");
        }
    }
}
