import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable } from 'rxjs';
import { tap } from 'rxjs/operators';

export interface OrderItem {
  id: string;         // Guid — potreban za brisanje
  tourId: string;
  tourName: string;
  price: number;
  shoppingCartId: string;
}

export interface ShoppingCart {
  id: string;
  touristId: string;
  items: OrderItem[];
  totalPrice: number;
}

// Checkout vraća direktno niz tokena (ne wrapper objekat)
export interface TourPurchaseToken {
  id: string;
  touristId: string;
  tourId: string;
  tourName: string;
  price: number;
  purchasedAt: string;
}

@Injectable({ providedIn: 'root' })
export class CartService {
  private apiUrl = 'http://localhost:8080';
  private cartSubject = new BehaviorSubject<ShoppingCart | null>(null);
  public cart$ = this.cartSubject.asObservable();

  constructor(private http: HttpClient) {}

  // GET /shopping-cart/{touristId}
  getCart(touristId: string): Observable<ShoppingCart> {
    return this.http.get<ShoppingCart>(`${this.apiUrl}/shopping-cart/${touristId}`).pipe(
      tap(cart => this.cartSubject.next(cart))
    );
  }

  // POST /shopping-cart/{touristId}/items
  addToCart(touristId: string, item: { tourId: string; tourName: string; price: number }): Observable<ShoppingCart> {
    return this.http.post<ShoppingCart>(`${this.apiUrl}/shopping-cart/${touristId}/items`, item).pipe(
      tap(cart => this.cartSubject.next(cart))
    );
  }

  // DELETE /shopping-cart/{touristId}/items/{itemId}  <-- itemId je Guid, ne tourId!
  removeFromCart(touristId: string, itemId: string): Observable<ShoppingCart> {
    return this.http.delete<ShoppingCart>(`${this.apiUrl}/shopping-cart/${touristId}/items/${itemId}`).pipe(
      tap(cart => this.cartSubject.next(cart))
    );
  }

  // POST /checkout/{touristId}  — vraća TourPurchaseToken[] direktno
  checkout(touristId: string): Observable<TourPurchaseToken[]> {
    return this.http.post<TourPurchaseToken[]>(`${this.apiUrl}/checkout/${touristId}`, {});
  }

  // GET /checkout/{touristId}/purchases
  getPurchasedTours(touristId: string): Observable<TourPurchaseToken[]> {
    return this.http.get<TourPurchaseToken[]>(`${this.apiUrl}/checkout/${touristId}/purchases`);
  }

  // GET /checkout/{touristId}/has-purchased/{tourId}
  hasPurchased(touristId: string, tourId: string): Observable<{ hasPurchased: boolean }> {
    return this.http.get<{ hasPurchased: boolean }>(`${this.apiUrl}/checkout/${touristId}/has-purchased/${tourId}`);
  }

  get currentCart(): ShoppingCart | null {
    return this.cartSubject.value;
  }
}