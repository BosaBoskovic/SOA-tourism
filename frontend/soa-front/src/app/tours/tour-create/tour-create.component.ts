import { Component, ChangeDetectorRef, NgZone, AfterViewInit, OnDestroy, PLATFORM_ID, Inject } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TourService, Tour, KeyPoint } from '../../services/tour.service';
import { AuthService } from '../../auth/services/auth.service';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-tour-create',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, FormsModule],
  templateUrl: './tour-create.component.html',
  styleUrl: './tour-create.component.css'
})
export class TourCreateComponent implements AfterViewInit, OnDestroy {
  form: FormGroup;
  loading = false;
  error = '';
  tagsInput = '';
  createdTour: Tour | null = null;

  keyPoints: KeyPoint[] = [];
  kpForm: FormGroup;
  kpLoading = false;
  kpError = '';
  selectedLatLng: { lat: number; lng: number } | null = null;

  private map: any = null;
  private L: any = null;
  private tempMarker: any = null;
  private keyPointMarkers: { keyPointId: string; marker: any }[] = [];
  private routePolyline: any = null;

  constructor(
    private fb: FormBuilder,
    private tourService: TourService,
    private authService: AuthService,
    private router: Router,
    private cdr: ChangeDetectorRef,
    private zone: NgZone,
    @Inject(PLATFORM_ID) private platformId: Object
  ) {
    this.form = this.fb.group({
      name: ['', [Validators.required]],
      description: ['', [Validators.required]],
      difficulty: ['', [Validators.required]],
      tags: [[]]
    });

    this.kpForm = this.fb.group({
      name: ['', Validators.required],
      description: [''],
      imageUrl: ['']
    });
  }

  addTag(): void {
    const tag = this.tagsInput.trim();
    if (tag) {
      const current: string[] = this.form.value.tags || [];
      this.form.patchValue({ tags: [...current, tag] });
      this.tagsInput = '';
    }
  }

  removeTag(tag: string): void {
    const current: string[] = this.form.value.tags || [];
    this.form.patchValue({ tags: current.filter(t => t !== tag) });
  }

  onSubmit(): void {
    if (this.form.invalid) return;

    this.loading = true;
    this.error = '';

    const user = this.authService['currentUserSubject'].getValue();

    this.tourService.createTour({
      ...this.form.value,
      authorId: user.username
    }).subscribe({
      next: (tour) => {
        console.log('CREATED TOUR:', tour);

        this.zone.run(() => {
          this.createdTour = tour;
          this.loading = false;
          this.cdr.detectChanges();

          if (isPlatformBrowser(this.platformId)) {
            setTimeout(() => this.initMap(), 200);
          }

          this.loadKeyPoints();
        });
      },

      error: (err) => {
        console.error('CREATE TOUR ERROR:', err);

        this.zone.run(() => {
          this.error = err.error?.error || 'Greška pri kreiranju ture.';
          this.loading = false;
          this.cdr.detectChanges();
        });
      }
    });
  }

  ngAfterViewInit(): void {
    if (this.createdTour && isPlatformBrowser(this.platformId)) {
      setTimeout(() => this.initMap(), 200);
    }
  }

  ngOnDestroy(): void {
    if (this.map) this.map.remove();
  }

  private async initMap(): Promise<void> {
    if (this.map || !this.createdTour) return;

    const mapContainer = document.getElementById('tour-map');
    if (!mapContainer) {
      setTimeout(() => this.initMap(), 100);
      return;
    }

    this.L = await import('leaflet');

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

    if (this.keyPoints.length > 0) {
      this.renderKeyPointMarkers();
    }
  }

  private renderKeyPointMarkers(): void {
    if (!this.map || !this.L) return;

    this.keyPointMarkers.forEach(item => {
      this.map.removeLayer(item.marker);
    });
    this.keyPointMarkers = [];

    if (this.routePolyline) {
      this.map.removeLayer(this.routePolyline);
      this.routePolyline = null;
    }

    const sortedKeyPoints = [...this.keyPoints].sort((a, b) => a.order - b.order);

    sortedKeyPoints.forEach((kp) => {
      const marker = this.L.marker([kp.latitude, kp.longitude])
        .addTo(this.map)
        .bindPopup(`<b>${kp.order}. ${kp.name}</b><br>${kp.description}`);

      if (kp.id) {
        this.keyPointMarkers.push({
          keyPointId: kp.id,
          marker: marker
        });
      }
    });

    if (sortedKeyPoints.length >= 2) {
      this.drawRoutePolyline(sortedKeyPoints);
    } else if (sortedKeyPoints.length === 1) {
      this.map.setView(
        [sortedKeyPoints[0].latitude, sortedKeyPoints[0].longitude],
        13
      );
    }
  }

  loadKeyPoints(): void {
    if (!this.createdTour) return;

    this.tourService.getKeyPointsByTour(this.createdTour.id).subscribe({
      next: (kps) => {
        this.zone.run(() => {
          this.keyPoints = kps;
          this.cdr.detectChanges();

          if (this.map) this.renderKeyPointMarkers();
        });
      }
    });
  }

  onKeyPointImageSelected(event: Event): void {
    const input = event.target as HTMLInputElement;

    if (!input.files || input.files.length === 0) {
      return;
    }

    const file = input.files[0];
    const reader = new FileReader();

    reader.onload = () => {
      this.kpForm.patchValue({
        imageUrl: reader.result as string
      });

      this.cdr.detectChanges();
    };

    reader.readAsDataURL(file);
  }

  addKeyPoint(): void {
    if (!this.createdTour) return;
    if (!this.selectedLatLng || this.kpForm.invalid) {
      this.kpError = 'Izaberite lokaciju na mapi i popunite naziv.';
      return;
    }

    this.kpLoading = true;
    this.kpError = '';

    const newKeyPoint = {
      tourId: this.createdTour.id,
      name: this.kpForm.value.name,
      description: this.kpForm.value.description,
      latitude: this.selectedLatLng.lat,
      longitude: this.selectedLatLng.lng,
      imageUrl: this.kpForm.value.imageUrl,
      order: this.keyPoints.length + 1
    };

    const nextKeyPoints = [...this.keyPoints, newKeyPoint];

    this.calculateLengthKm(nextKeyPoints)
      .then((lengthKm) => {
        this.tourService.createKeyPoint({
          ...newKeyPoint,
          lengthKm
        }).subscribe({
          next: () => {
            this.kpForm.reset();
            this.selectedLatLng = null;

            if (this.tempMarker) {
              this.map.removeLayer(this.tempMarker);
              this.tempMarker = null;
            }

            this.kpLoading = false;
            this.loadKeyPoints();
          },
          error: () => {
            this.kpError = 'Greška pri dodavanju ključne tačke.';
            this.kpLoading = false;
          }
        });
      })
      .catch(() => {
        this.kpError = 'Greška pri izračunavanju dužine ture.';
        this.kpLoading = false;
      });
  }

  deleteKeyPoint(kp: KeyPoint): void {
    if (!this.createdTour) return;

    const nextKeyPoints = this.keyPoints.filter(k => k.id !== kp.id);

    this.calculateLengthKm(nextKeyPoints)
      .then((lengthKm) => {
        this.tourService.deleteKeyPoint(kp.id!, lengthKm).subscribe({
          next: () => {
            this.loadKeyPoints();
          }
        });
      })
      .catch(() => {
        this.kpError = 'Greška pri izračunavanju dužine ture.';
      });
  }

  finishSetup(): void {
    if (!this.createdTour) return;
    this.router.navigate(['/tours', this.createdTour.id]);
  }

  private async calculateLengthKm(points: Array<Pick<KeyPoint, 'latitude' | 'longitude' | 'order'>>): Promise<number> {
    if (points.length < 2) {
      return 0;
    }

    const data = await this.fetchRouteData(points);
    const legs = data?.routes?.[0]?.legs ?? [];
    if (legs.length > 0) {
      const meters = legs.reduce((sum: number, leg: { distance?: number }) => sum + (leg.distance ?? 0), 0);
      return meters / 1000;
    }

    const meters = data?.routes?.[0]?.distance ?? 0;
    return meters / 1000;
  }

  private async drawRoutePolyline(points: Array<Pick<KeyPoint, 'latitude' | 'longitude' | 'order'>>): Promise<void> {
    try {
      const data = await this.fetchRouteData(points, true);
      const coords = data?.routes?.[0]?.geometry?.coordinates ?? [];
      if (coords.length === 0) {
        return;
      }

      const latLngs = coords.map((coord: [number, number]) => [coord[1], coord[0]]);

      this.routePolyline = this.L.polyline(latLngs, {
        color: '#2563eb',
        weight: 5,
        opacity: 0.85
      }).addTo(this.map);

      this.map.fitBounds(this.routePolyline.getBounds(), {
        padding: [40, 40]
      });
    } catch {
      const fallback = points
        .sort((a, b) => a.order - b.order)
        .map(kp => [kp.latitude, kp.longitude]);

      this.routePolyline = this.L.polyline(fallback, {
        color: '#2563eb',
        weight: 5,
        opacity: 0.85
      }).addTo(this.map);
    }
  }

  private async fetchRouteData(
    points: Array<Pick<KeyPoint, 'latitude' | 'longitude' | 'order'>>,
    withGeometry = false
  ): Promise<any> {
    const sorted = [...points].sort((a, b) => a.order - b.order);
    const coords = sorted.map((kp) => `${kp.longitude},${kp.latitude}`).join(';');
    const overview = withGeometry ? 'full' : 'false';
    const geometries = withGeometry ? '&geometries=geojson' : '';
    const url = `https://router.project-osrm.org/route/v1/foot/${coords}?overview=${overview}${geometries}`;

    const response = await fetch(url);
    if (!response.ok) {
      throw new Error('Routing request failed');
    }

    return response.json();
  }
}