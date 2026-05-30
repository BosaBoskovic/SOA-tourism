import { Component, OnInit, ChangeDetectorRef, NgZone } from '@angular/core';
import { CommonModule, SlicePipe, DatePipe } from '@angular/common';
import { RouterLink } from '@angular/router';
import { CartService, ShoppingCart, TourPurchaseToken, OrderItem } from '../services/cart.service';
import { AuthService } from '../auth/services/auth.service';
import { TopNavComponent } from '../shared/top-nav/top-nav.component';
import { TourService } from '../services/tour.service'; 
import { forkJoin, of } from 'rxjs';                     
import { catchError } from 'rxjs/operators'; 


@Component({
  selector: 'app-shopping-cart',
  standalone: true,
  imports: [CommonModule, RouterLink, TopNavComponent, SlicePipe, DatePipe],
  templateUrl: './shopping-cart.component.html',
  styleUrl: './shopping-cart.component.css'
})
export class ShoppingCartComponent implements OnInit {
  cart: ShoppingCart | null = null;
  purchasedTokens: TourPurchaseToken[] = [];
  currentUser: any;
  loading = false;
  checkoutLoading = false;
  error = '';
  successMessage = '';
  showTokens = false;
  archivedTourIds: Set<string> = new Set();

  constructor(
    private cartService: CartService,
    private authService: AuthService,
    private cdr: ChangeDetectorRef,
    private zone: NgZone,
    private tourService: TourService
  ) {}

  ngOnInit(): void {
    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;
      if (user?.username) {
        this.loadCart(user.username);
      }
    });
  }

  loadCart(touristId: string): void {
    this.loading = true;
    this.error = '';

    this.cartService.getCart(touristId).subscribe({
      next: (cart) => {
        this.zone.run(() => {
          this.cart = cart;
          this.checkForArchivedTours(cart);
          this.loading = false;
          this.cdr.detectChanges();
        });
      },
      error: (err) => {
        this.zone.run(() => {
          this.error = err.error?.error || 'Greška pri učitavanju korpe.';
          this.loading = false;
          this.cdr.detectChanges();
        });
      }
    });
  }

  checkForArchivedTours(cart: ShoppingCart): void {
    if (!cart.items?.length) {
      this.zone.run(() => {
        this.loading = false;
        this.cdr.detectChanges();
      });
      return;
    }
  
    const checks$ = cart.items.map(item =>
      this.tourService.getTourById(item.tourId).pipe(
        catchError(() => of(null))
      )
    );
  
    forkJoin(checks$).subscribe(tours => {
      this.zone.run(() => {
        this.archivedTourIds = new Set(
          tours
            .filter(t => t?.status === 'archived')
            .map(t => t!.id)
        );
        this.loading = false;
        this.cdr.detectChanges();
      });
    });
  }
  
  isArchived(item: OrderItem): boolean {
    return this.archivedTourIds.has(item.tourId);
  }
  
  get hasArchivedItems(): boolean {
    return this.archivedTourIds.size > 0;
  }

  // Prima item.id (Guid) — ne tourId!
  removeItem(itemId: string): void {
    if (!this.currentUser?.username) return;

    this.cartService.removeFromCart(this.currentUser.username, itemId).subscribe({
      next: (cart) => {
        this.zone.run(() => {
          this.cart = cart;
          this.cdr.detectChanges();
        });
      },
      error: (err) => {
        this.zone.run(() => {
          this.error = err.error?.error || 'Greška pri uklanjanju stavke.';
          this.cdr.detectChanges();
        });
      }
    });
  }

  // Checkout vraća TourPurchaseToken[] direktno (bez wrapper-a)
  checkout(): void {
    if (!this.currentUser?.username) return;
    if (!this.cart?.items?.length) return;

    this.checkoutLoading = true;
    this.error = '';
    this.successMessage = '';

    this.cartService.checkout(this.currentUser.username).subscribe({
      next: (tokens) => {
        this.zone.run(() => {
          this.purchasedTokens = tokens;
          this.successMessage = `Kupovina uspešna! Dobili ste ${tokens.length} token(a).`;
          this.showTokens = true;
          this.cart = {
            id: this.cart!.id,
            touristId: this.currentUser.username,
            items: [],
            totalPrice: 0
          };
          this.checkoutLoading = false;
          this.cdr.detectChanges();
        });
      },
      error: (err) => {
        this.zone.run(() => {
         const rawError =
  err.error?.error ||
  err.error?.message ||
  err.error?.detail ||
  err.message ||
  '';

if (rawError.includes('soa-tours') || rawError.includes('Name or service not known')) {
  this.error = 'Tura trenutno nije dostupna za kupovinu. Pokušajte kasnije.';
} else if (rawError.includes('not available') || rawError.includes('nije dostupna')) {
  this.error = 'Tura više nije dostupna za kupovinu.';
} else {
  this.error = 'Kupovina nije uspela. Sistem je poništio započetu transakciju.';
}
          this.checkoutLoading = false;
          this.cdr.detectChanges();
        });
      }
    });
  }

  loadPurchasedTokens(): void {
    if (!this.currentUser?.username) return;

    this.cartService.getPurchasedTours(this.currentUser.username).subscribe({
      next: (tokens) => {
        this.zone.run(() => {
          this.purchasedTokens = tokens;
          this.showTokens = true;
          this.cdr.detectChanges();
        });
      },
      error: (err) => {
        this.zone.run(() => {
          this.error = err.error?.error || 'Greška pri učitavanju kupljenih tura.';
          this.cdr.detectChanges();
        });
      }
    });
  }

  get itemCount(): number {
    return this.cart?.items?.length ?? 0;
  }
}