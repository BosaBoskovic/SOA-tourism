import { Component, OnInit, AfterViewInit, OnDestroy, PLATFORM_ID, Inject } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TourService, Tour, KeyPoint } from '../../services/tour.service';
import { ReviewService, Review } from '../../services/review.service';
import { AuthService } from '../../auth/services/auth.service';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';

@Component({
  selector: 'app-tour-detail',
  standalone: true,
  imports: [CommonModule, RouterLink, ReactiveFormsModule],
  templateUrl: './tour-detail.component.html',
  styleUrl: './tour-detail.component.css'
})
export class TourDetailComponent implements OnInit, AfterViewInit, OnDestroy {
  tour: Tour | null = null;
  keyPoints: KeyPoint[] = [];
  reviews: Review[] = [];
  currentUser: any;
  loading = true;
  error = '';

  private map: any = null;
  private L: any = null;
  private tempMarker: any = null;
  selectedLatLng: { lat: number; lng: number } | null = null;

  kpForm: FormGroup;
  kpLoading = false;
  kpError = '';

  reviewForm: FormGroup;
  reviewLoading = false;
  reviewError = '';
  reviewSuccess = '';

  constructor(
    private route: ActivatedRoute,
    private tourService: TourService,
    private reviewService: ReviewService,
    private authService: AuthService,
    private fb: FormBuilder,
    @Inject(PLATFORM_ID) private platformId: Object
  ) {
    this.kpForm = this.fb.group({
      name: ['', Validators.required],
      description: [''],
      imageUrl: ['']
    });

    this.reviewForm = this.fb.group({
      rating: [5, [Validators.required, Validators.min(1), Validators.max(5)]],
      comment: ['', Validators.required],
      tourVisitDate: ['', Validators.required]
    });
  }

  ngOnInit(): void {
    this.authService.currentUser$.subscribe(u => this.currentUser = u);
    const id = this.route.snapshot.paramMap.get('id')!;
    this.tourService.getTourById(id).subscribe({
      next: (tour) => {
        this.tour = tour;
        this.loading = false;
        this.loadKeyPoints(id);
        this.loadReviews(id);
      },
      error: () => { this.error = 'Tura nije pronađena.'; this.loading = false; }
    });
  }

  async ngAfterViewInit(): Promise<void> {
    if (isPlatformBrowser(this.platformId)) {
      await this.initMap();
    }
  }

  ngOnDestroy(): void {
    if (this.map) this.map.remove();
  }

  private async initMap(): Promise<void> {
    // Dinamički import - izvršava se SAMO u browseru
    this.L = await import('leaflet');

    // Fix za ikonice
    const iconDefault = this.L.icon({
      iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
      iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
      shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
      iconSize: [25, 41],
      iconAnchor: [12, 41],
      popupAnchor: [1, -34],
      shadowSize: [41, 41]
    });
    this.L.Marker.prototype.options.icon = iconDefault;

    this.map = this.L.map('tour-map').setView([44.0, 21.0], 7);
    this.L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© OpenStreetMap contributors'
    }).addTo(this.map);

    this.map.on('click', (e: any) => {
      if (this.tempMarker) this.map.removeLayer(this.tempMarker);
      this.selectedLatLng = { lat: e.latlng.lat, lng: e.latlng.lng };
      this.tempMarker = this.L.marker(e.latlng).addTo(this.map)
        .bindPopup('Nova ključna tačka').openPopup();
    });

    // Ako su ključne tačke već učitane prije mape, nacrtaj ih
    if (this.keyPoints.length > 0) {
      this.renderKeyPointMarkers();
    }
  }

  private renderKeyPointMarkers(): void {
    if (!this.map || !this.L) return;
    this.keyPoints.forEach((kp) => {
      this.L.marker([kp.latitude, kp.longitude])
        .addTo(this.map)
        .bindPopup(`<b>${kp.name}</b><br>${kp.description}`);
    });
  }

  loadKeyPoints(tourId: string): void {
    this.tourService.getKeyPointsByTour(tourId).subscribe({
      next: (kps) => {
        this.keyPoints = kps;
        // Mapa možda još nije inicijalizovana, renderKeyPointMarkers() čeka na map
        if (this.map) this.renderKeyPointMarkers();
      }
    });
  }

  loadReviews(tourId: string): void {
    this.reviewService.getReviewsByTour(tourId).subscribe({
      next: (reviews) => this.reviews = reviews
    });
  }

  addKeyPoint(): void {
    if (!this.selectedLatLng || this.kpForm.invalid) {
      this.kpError = 'Izaberite lokaciju na mapi i popunite naziv.';
      return;
    }
    this.kpLoading = true;
    this.kpError = '';

    this.tourService.createKeyPoint({
      tourId: this.tour!.id,
      name: this.kpForm.value.name,
      description: this.kpForm.value.description,
      latitude: this.selectedLatLng.lat,
      longitude: this.selectedLatLng.lng,
      imageUrl: this.kpForm.value.imageUrl,
      order: this.keyPoints.length + 1
    }).subscribe({
      next: (kp) => {
        this.keyPoints.push(kp);
        if (this.map && this.L) {
          this.L.marker([kp.latitude, kp.longitude]).addTo(this.map)
            .bindPopup(`<b>${kp.name}</b>`);
        }
        this.kpForm.reset();
        this.selectedLatLng = null;
        if (this.tempMarker) { this.map.removeLayer(this.tempMarker); this.tempMarker = null; }
        this.kpLoading = false;
      },
      error: () => { this.kpError = 'Greška pri dodavanju ključne tačke.'; this.kpLoading = false; }
    });
  }

  deleteKeyPoint(kp: KeyPoint): void {
    this.tourService.deleteKeyPoint(kp.id!).subscribe({
      next: () => this.keyPoints = this.keyPoints.filter(k => k.id !== kp.id)
    });
  }

  submitReview(): void {
    if (this.reviewForm.invalid) return;
    this.reviewLoading = true;
    this.reviewError = '';

    this.reviewService.createReview({
      tourId: this.tour!.id,
      touristId: this.currentUser.username,
      touristName: this.currentUser.username,
      rating: this.reviewForm.value.rating,
      comment: this.reviewForm.value.comment,
      images: [],
      tourVisitDate: this.reviewForm.value.tourVisitDate
    }).subscribe({
      next: (review) => {
        this.reviews.push(review);
        this.reviewForm.reset({ rating: 5 });
        this.reviewSuccess = 'Recenzija uspješno dodata!';
        this.reviewLoading = false;
      },
      error: (err) => {
        this.reviewError = err.error?.error || 'Greška pri slanju recenzije.';
        this.reviewLoading = false;
      }
    });
  }

  isGuide(): boolean { return this.currentUser?.role === 'guide'; }
  isTourist(): boolean { return this.currentUser?.role === 'tourist'; }
  isOwner(): boolean { return this.tour?.authorId === this.currentUser?.username; }
  starsArray(n: number): number[] { return Array(n).fill(0); }
}