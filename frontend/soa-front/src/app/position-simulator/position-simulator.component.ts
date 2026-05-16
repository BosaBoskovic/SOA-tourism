import { Component, AfterViewInit, Inject, PLATFORM_ID, ChangeDetectorRef } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { AuthService } from '../auth/services/auth.service';
import { PositionService, TouristPosition } from '../services/position.service';

@Component({
  selector: 'app-position-simulator',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './position-simulator.component.html',
  styleUrl: './position-simulator.component.css'
})
export class PositionSimulatorComponent implements AfterViewInit {
  private map: any = null;
  private L: any = null;
  private marker: any = null;

  currentUser: any;
  currentPosition: TouristPosition | null = null;

  constructor(
    private authService: AuthService,
    private positionService: PositionService,
    private cdr: ChangeDetectorRef,
    @Inject(PLATFORM_ID) private platformId: Object
  ) {
    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;
      this.currentPosition = this.positionService.getPosition();
    });
  }

  async ngAfterViewInit(): Promise<void> {
    if (isPlatformBrowser(this.platformId)) {
      await this.initMap();
    }
  }

  private async initMap(): Promise<void> {
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

    const startLat = this.currentPosition?.latitude ?? 44.0;
    const startLng = this.currentPosition?.longitude ?? 21.0;
    const zoom = this.currentPosition ? 13 : 7;

    this.map = this.L.map('position-map').setView([startLat, startLng], zoom);

    this.L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© OpenStreetMap contributors'
    }).addTo(this.map);

    if (this.currentPosition) {
      this.setMarker(
        this.currentPosition.latitude,
        this.currentPosition.longitude
      );
    }

    this.map.on('click', (e: any) => {
      const lat = e.latlng.lat;
      const lng = e.latlng.lng;

      this.currentPosition = {
        touristId: this.currentUser.username,
        latitude: lat,
        longitude: lng
      };

      this.positionService.savePosition(this.currentPosition);
      this.setMarker(lat, lng);
      this.cdr.detectChanges();
    });
  }

  private setMarker(lat: number, lng: number): void {
    if (this.marker) {
      this.map.removeLayer(this.marker);
    }

    this.marker = this.L.marker([lat, lng])
      .addTo(this.map)
      .bindPopup('Moja trenutna lokacija')
      .openPopup();
  }
}