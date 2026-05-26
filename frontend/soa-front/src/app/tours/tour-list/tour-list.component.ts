import { Component, OnInit, ChangeDetectorRef, NgZone } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { TourService, Tour, TourPreview } from '../../services/tour.service';
import { AuthService } from '../../auth/services/auth.service';
import { ReviewFormComponent } from '../../reviews/review-form/review-form.component';
import { ReviewService } from '../../services/review.service';
import { CartService } from '../../services/cart.service';

@Component({
  selector: 'app-tour-list',
  standalone: true,
  imports: [CommonModule, RouterLink, ReviewFormComponent],
  templateUrl: './tour-list.component.html',
  styleUrl: './tour-list.component.css'
})
export class TourListComponent implements OnInit {
  tours: Array<Tour | TourPreview> = [];
  loading = false;
  error = '';
  currentUser: any;
  selectedTourForReview: Tour | TourPreview | null = null;
  cartTourIds: Set<string> = new Set();
  purchasedTourIds: Set<string> = new Set();

  constructor(
    private tourService: TourService,
    private authService: AuthService,
    private cdr: ChangeDetectorRef,
    private zone: NgZone,
    private reviewService: ReviewService,
    private cartService: CartService,
  ) {}

  ngOnInit(): void {
    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;

      if (user?.role === 'guide') {
        this.loadToursByAuthor(user.username);
      }

      if (user?.role === 'tourist') {
        this.loadAllTours();
        this.loadCartState(user.username);
        this.loadPurchasedTours(user.username);
      }
    });

    this.cartService.cart$.subscribe(cart => {
      this.cartTourIds = new Set(cart?.items?.map(i => i.tourId) ?? []);
      this.cdr.detectChanges();
    });
  }

  loadCartState(touristId: string): void {
    this.cartService.getCart(touristId).subscribe({
      next: (cart) => {
        this.cartTourIds = new Set(cart.items?.map(i => i.tourId) ?? []);
        this.cdr.detectChanges();
      },
      error: () => { this.cartTourIds = new Set(); }
    });
  }

  loadPurchasedTours(touristId: string): void {
    this.cartService.getPurchasedTours(touristId).subscribe({
      next: (tokens) => {
        this.purchasedTourIds = new Set(tokens.map(t => t.tourId));
        this.cdr.detectChanges();
      },
      error: () => { this.purchasedTourIds = new Set(); }
    });
  }

  addToCart(tour: Tour | TourPreview): void {
    if (!this.currentUser?.username) return;

    this.cartService.addToCart(this.currentUser.username, {
      tourId: tour.id,
      tourName: tour.name,
      price: tour.price ?? 0
    }).subscribe({
      next: () => {
        this.zone.run(() => {
          this.cartTourIds = new Set([...this.cartTourIds, tour.id]);
          this.cdr.detectChanges();
        });
      },
      error: (err) => {
        alert(err.error?.error || 'Greška pri dodavanju u korpu.');
      }
    });
  }

  isInCart(tour: Tour | TourPreview): boolean {
    return this.cartTourIds.has(tour.id);
  }

  isPurchased(tour: Tour | TourPreview): boolean {
    return this.purchasedTourIds.has(tour.id);
  }

  loadToursByAuthor(authorId: string): void {
    this.loading = true;
    this.error = '';

    this.tourService.getToursByAuthor(authorId).subscribe({
      next: (tours) => {
        this.zone.run(() => {
          this.tours = tours;
          this.loading = false;
          this.cdr.detectChanges();
        });
      },
      error: (err) => {
        this.zone.run(() => {
          this.error = err.error?.error || 'Greška pri učitavanju tura.';
          this.loading = false;
          this.cdr.detectChanges();
        });
      }
    });
  }

  loadAllTours(): void {
    this.loading = true;
    this.error = '';

    this.tourService.getAllTours().subscribe({
      next: (tours) => {
        this.zone.run(() => {
          this.tours = tours;
          this.loading = false;
          this.cdr.detectChanges();
        });
      },
      error: (err) => {
        this.zone.run(() => {
          this.error = err.error?.error || 'Greška pri učitavanju svih tura.';
          this.loading = false;
          this.cdr.detectChanges();
        });
      }
    });
  }

  openReviewForm(tour: Tour | TourPreview): void {
    this.selectedTourForReview = tour;
  }

  submitReview(review: any): void {
    if (!this.selectedTourForReview) return;

    const request = {
      tourId: this.selectedTourForReview.id,
      touristId: this.currentUser?.username,
      touristName: this.currentUser?.username,
      rating: Number(review.rating),
      comment: review.comment,
      images: review.images,
      tourVisitDate: review.visitDate
    };

    this.reviewService.createReview(request).subscribe({
      next: () => {
        alert('Recenzija je uspešno dodata.');
        this.selectedTourForReview = null;
      },
      error: (err) => {
        alert(err.error?.error || 'Greška pri dodavanju recenzije.');
      }
    });
  }

  publishTour(tour: Tour): void {
    this.tourService.publishTour(tour.id).subscribe({
      next: (updatedTour) => {
        this.zone.run(() => {
          tour.status = updatedTour.status;
          this.cdr.detectChanges();
        });
        alert('Tura je uspešno objavljena.');
      },
      error: (err) => {
        alert(err.error?.error || 'Greška pri objavljivanju ture.');
      }
    });
  }

  archiveTour(tour: Tour): void {
    this.tourService.archiveTour(tour.id).subscribe({
      next: (updatedTour) => {
        this.zone.run(() => {
          tour.status = updatedTour.status;
          this.cdr.detectChanges();
        });
        alert('Tura je uspešno arhivirana.');
      },
      error: (err) => {
        alert(err.error?.error || 'Greška pri arhiviranju ture.');
      }
    });
  }

  activateTour(tour: Tour): void {
    this.tourService.activateTour(tour.id).subscribe({
      next: (updatedTour) => {
        this.zone.run(() => {
          tour.status = updatedTour.status;
          this.cdr.detectChanges();
        });
        alert('Tura je uspešno aktivirana.');
      },
      error: (err) => {
        alert(err.error?.error || 'Greška pri aktiviranju ture.');
      }
    });
  }

  difficultyLabel(d: string): string {
    return { easy: 'Lako', medium: 'Srednje', hard: 'Teško' }[d] ?? d;
  }

  hasStatus(tour: Tour | TourPreview): tour is Tour {
    return (tour as Tour).status !== undefined;
  }
}