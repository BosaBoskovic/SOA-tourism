import { Component, OnInit, ChangeDetectorRef, NgZone } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { FormBuilder, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { TourService, Tour, TourPreview, TourAnalytics, TourSearchParams } from '../../services/tour.service';
import { AuthService } from '../../auth/services/auth.service';
import { ReviewFormComponent } from '../../reviews/review-form/review-form.component';
import { ReviewService } from '../../services/review.service';
import { CartService } from '../../services/cart.service';
import { ToastService } from '../../shared/toast/toast.service';
import { ConfirmDialogService } from '../../shared/confirm-dialog/confirm-dialog.service';

@Component({
  selector: 'app-tour-list',
  standalone: true,
  imports: [CommonModule, RouterLink, ReactiveFormsModule, ReviewFormComponent],
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

  filterForm: FormGroup;
  analytics: Record<string, TourAnalytics> = {};

  constructor(
    private tourService: TourService,
    private authService: AuthService,
    private cdr: ChangeDetectorRef,
    private zone: NgZone,
    private reviewService: ReviewService,
    private cartService: CartService,
    private toastService: ToastService,
    private confirmDialogService: ConfirmDialogService,
    private fb: FormBuilder,
  ) {
    this.filterForm = this.fb.group({
      difficulty: [''],
      minPrice: [null],
      maxPrice: [null],
      sortBy: ['name'],
      sortDir: ['asc'],
    });
  }

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
        this.toastService.error(err.error?.error || 'Greška pri dodavanju u korpu.');
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
          this.loadAnalytics(tours);
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

  private loadAnalytics(tours: Array<Tour | TourPreview>): void {
    const tourIds = tours.map(t => t.id);
    this.tourService.getAnalytics(tourIds).subscribe({
      next: (results) => {
        this.zone.run(() => {
          this.analytics = {};
          for (const a of results) {
            this.analytics[a.tourId] = a;
          }
          this.cdr.detectChanges();
        });
      },
      error: () => {
        // Non-fatal: the tour list still works without stats.
      }
    });
  }

  onFilterChange(): void {
    if (this.currentUser?.role === 'tourist') {
      this.loadAllTours();
    }
  }

  resetFilters(): void {
    this.filterForm.reset({ difficulty: '', minPrice: null, maxPrice: null, sortBy: 'name', sortDir: 'asc' });
    this.onFilterChange();
  }

  loadAllTours(): void {
    this.loading = true;
    this.error = '';

    const raw = this.filterForm.value;
    const params: TourSearchParams = {
      difficulty: raw.difficulty || undefined,
      minPrice: raw.minPrice ?? undefined,
      maxPrice: raw.maxPrice ?? undefined,
      sortBy: raw.sortBy || undefined,
      sortDir: raw.sortDir || undefined,
    };

    this.tourService.getAllTours(params).subscribe({
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
        this.toastService.success('Recenzija je uspešno dodata.');
        this.selectedTourForReview = null;
      },
      error: (err) => {
        this.toastService.error(err.error?.error || 'Greška pri dodavanju recenzije.');
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
        this.toastService.success('Tura je uspešno objavljena.');
      },
      error: (err) => {
        this.toastService.error(err.error?.error || 'Greška pri objavljivanju ture.');
      }
    });
  }

  async archiveTour(tour: Tour): Promise<void> {
    const confirmed = await this.confirmDialogService.confirm(
      `Da li ste sigurni da želite da arhivirate turu "${tour.name}"? Više neće biti vidljiva turistima.`
    );
    if (!confirmed) return;

    this.tourService.archiveTour(tour.id).subscribe({
      next: (updatedTour) => {
        this.zone.run(() => {
          tour.status = updatedTour.status;
          this.cdr.detectChanges();
        });
        this.toastService.success('Tura je uspešno arhivirana.');
      },
      error: (err) => {
        this.toastService.error(err.error?.error || 'Greška pri arhiviranju ture.');
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
        this.toastService.success('Tura je uspešno aktivirana.');
      },
      error: (err) => {
        this.toastService.error(err.error?.error || 'Greška pri aktiviranju ture.');
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