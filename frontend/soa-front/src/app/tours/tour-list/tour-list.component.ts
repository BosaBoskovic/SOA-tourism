import { Component, OnInit, ChangeDetectorRef, NgZone } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { TourService, Tour } from '../../services/tour.service';
import { AuthService } from '../../auth/services/auth.service';
import { ReviewFormComponent } from '../../reviews/review-form/review-form.component';
import { ReviewService } from '../../services/review.service';

@Component({
  selector: 'app-tour-list',
  standalone: true,
  imports: [CommonModule, RouterLink, ReviewFormComponent],
  templateUrl: './tour-list.component.html',
  styleUrl: './tour-list.component.css'
})
export class TourListComponent implements OnInit {
  tours: Tour[] = [];
  loading = false;
  error = '';
  currentUser: any;

  selectedTourForReview: Tour | null = null;

  constructor(
    private tourService: TourService,
    private authService: AuthService,
    private cdr: ChangeDetectorRef,
    private zone: NgZone,
    private reviewService: ReviewService,
  ) {}

  ngOnInit(): void {
    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;

      if (user?.role === 'guide') {
        this.loadToursByAuthor(user.username);
      }

      if (user?.role === 'tourist') {
        this.loadAllTours();
      }
    });
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
          this.tours = tours.filter(tour => tour.status === 'published');
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

  openReviewForm(tour: Tour): void {
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

  difficultyLabel(d: string): string {
    return { easy: 'Lako', medium: 'Srednje', hard: 'Teško' }[d] ?? d;
  }
}