import {
    Component, OnInit, OnDestroy, AfterViewInit,
    Inject, PLATFORM_ID, ChangeDetectorRef, NgZone
  } from '@angular/core';
  import { CommonModule, isPlatformBrowser, DecimalPipe, DatePipe } from '@angular/common';
  import { ActivatedRoute, Router, RouterLink } from '@angular/router';
  import { ExecutionService, TourExecution, CheckKeyPointResponse } from '../services/execution.service';
  import { TourService, KeyPoint } from '../services/tour.service';
  import { PositionService, TouristPosition } from '../services/position.service';
  import { AuthService } from '../auth/services/auth.service';
  import { TopNavComponent } from '../shared/top-nav/top-nav.component';
  
  @Component({
    selector: 'app-tour-execution',
    standalone: true,
    imports: [CommonModule, RouterLink, TopNavComponent, DecimalPipe, DatePipe],
    templateUrl: './tour-execution.component.html',
    styleUrl: './tour-execution.component.css'
  })
  export class TourExecutionComponent implements OnInit, AfterViewInit, OnDestroy {
    execution: TourExecution | null = null;
    keyPoints: KeyPoint[] = [];
    currentUser: any;
    tourId: string = '';
  
    loading = true;
    error = '';
    lastCheckMessage = '';
    lastReachedKeyPoint: string | null = null;
    checkingNow = false;
  
    // Position panel state
    showPositionPanel = false;
    pendingPosition: TouristPosition | null = null;
    currentSimPosition: TouristPosition | null = null;
  
    // Countdown
    countdownSeconds = 10;
    private countdownInterval: any = null;
  
    private map: any = null;
    private L: any = null;
    private playerMarker: any = null;
    private keyPointMarkers: any[] = [];;
    private pollingInterval: any = null;
  
    // Mini simulator map
    private simMap: any = null;
    private simMarker: any = null;
    // Mini position panel map
    private miniMap: any = null;
    private miniMarker: any = null;
  
    constructor(
      private route: ActivatedRoute,
      private router: Router,
      private executionService: ExecutionService,
      private tourService: TourService,
      private positionService: PositionService,
      private authService: AuthService,
      private cdr: ChangeDetectorRef,
      private zone: NgZone,
      @Inject(PLATFORM_ID) private platformId: Object
    ) {}
  
    ngOnInit(): void {
      this.tourId = this.route.snapshot.paramMap.get('tourId')!;
      const executionId = this.route.snapshot.paramMap.get('executionId');
  
      this.authService.currentUser$.subscribe(user => {
        this.currentUser = user;
      });
  
      this.currentSimPosition = this.positionService.getPosition();
  
      if (executionId) {
        this.executionService.getById(executionId).subscribe({
          next: (exec) => {
            this.zone.run(() => {
              this.execution = exec;
              this.loading = false;
              this.loadKeyPoints();
              this.cdr.detectChanges();
            });
          },
          error: () => {
            this.zone.run(() => {
              this.error = 'Sesija nije pronađena.';
              this.loading = false;
              this.cdr.detectChanges();
            });
          }
        });
      } else {
        this.startExecution();
      }
    }
  
    async ngAfterViewInit(): Promise<void> {
      if (isPlatformBrowser(this.platformId)) {
        await this.initMap();
      }
    }
  
    ngOnDestroy(): void {
      this.stopPolling();
      this.stopCountdown();
      if (this.map) this.map.remove();
      if (this.simMap) this.simMap.remove();
      if (this.miniMap) this.miniMap.remove();
    }
  
    // ---- Position Panel ----
  
    openPositionPanel(): void {
      this.showPositionPanel = true;
      this.pendingPosition = null;
      setTimeout(() => this.initMiniMap(), 200);
    }
  
    closePositionPanel(): void {
      this.showPositionPanel = false;
      this.pendingPosition = null;
      if (this.miniMap) { this.miniMap.remove(); this.miniMap = null; }
    }
  
    confirmPosition(): void {
      if (!this.pendingPosition) return;
      this.positionService.savePosition(this.pendingPosition);
      this.currentSimPosition = this.pendingPosition;
      this.closePositionPanel();
      this.error = '';
      this.startExecution();
    }
  
    private async initMiniMap(): Promise<void> {
      const el = document.getElementById('mini-position-map');
      if (!el) return;
  
      if (!this.L) this.L = await import('leaflet');
  
      const pos = this.positionService.getPosition();
      const lat = pos?.latitude ?? 44.0;
      const lng = pos?.longitude ?? 21.0;
  
      this.miniMap = this.L.map('mini-position-map').setView([lat, lng], pos ? 13 : 7);
      this.L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors'
      }).addTo(this.miniMap);
  
      if (pos) {
        this.miniMarker = this.L.marker([pos.latitude, pos.longitude]).addTo(this.miniMap);
      }
  
      this.miniMap.on('click', (e: any) => {
        const { lat, lng } = e.latlng;
        this.pendingPosition = { touristId: this.currentUser?.username ?? '', latitude: lat, longitude: lng };
        if (this.miniMarker) this.miniMap.removeLayer(this.miniMarker);
        this.miniMarker = this.L.marker([lat, lng]).addTo(this.miniMap).bindPopup('Moja lokacija').openPopup();
        this.cdr.detectChanges();
      });
    }
  
    // ---- Inline Simulator Sidebar ----
  
    private async initSimSidebar(): Promise<void> {
      await new Promise(r => setTimeout(r, 100));
      const el = document.getElementById('sim-mini-map');
      if (!el || this.simMap) return;
  
      const pos = this.positionService.getPosition();
      const lat = pos?.latitude ?? 44.0;
      const lng = pos?.longitude ?? 21.0;
  
      this.simMap = this.L.map('sim-mini-map').setView([lat, lng], 13);
      this.L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors'
      }).addTo(this.simMap);
  
      if (pos) {
        const simIcon = this.L.divIcon({
          html: '<div style="background:#4f7cff;border:3px solid white;border-radius:50%;width:16px;height:16px"></div>',
          iconSize: [16, 16], iconAnchor: [8, 8], className: ''
        });
        this.simMarker = this.L.marker([pos.latitude, pos.longitude], { icon: simIcon }).addTo(this.simMap);
      }
  
      this.simMap.on('click', (e: any) => {
        const { lat, lng } = e.latlng;
        const newPos: TouristPosition = { touristId: this.currentUser?.username ?? '', latitude: lat, longitude: lng };
        this.positionService.savePosition(newPos);
        this.currentSimPosition = newPos;
  
        const simIcon = this.L.divIcon({
          html: '<div style="background:#4f7cff;border:3px solid white;border-radius:50%;width:16px;height:16px"></div>',
          iconSize: [16, 16], iconAnchor: [8, 8], className: ''
        });
        if (this.simMarker) this.simMap.removeLayer(this.simMarker);
        this.simMarker = this.L.marker([lat, lng], { icon: simIcon }).addTo(this.simMap).bindPopup('Moja pozicija').openPopup();
  
        // Also update main map
        this.updatePlayerMarker(lat, lng);
        this.cdr.detectChanges();
      });
    }
  
    // ---- Execution ----
  
    private startExecution(): void {
      const position = this.positionService.getPosition();
  
      if (!position) {
        this.error = 'Niste postavili svoju lokaciju. Postavite je ovdje:';
        this.loading = false;
        this.cdr.detectChanges();
        return;
      }
  
      const req = {
        touristId: this.currentUser.username,
        tourId: this.tourId,
        latitude: position.latitude,
        longitude: position.longitude
      };
  
      this.executionService.start(req).subscribe({
        next: (exec) => {
          this.zone.run(() => {
            this.execution = exec;
            this.loading = false;
            this.loadKeyPoints();
            this.cdr.detectChanges();
          });
        },
        error: (err) => {
          this.zone.run(() => {
            this.error = err.error?.error || 'Greška pri pokretanju ture.';
            this.loading = false;
            this.cdr.detectChanges();
          });
        }
      });
    }
  
    private loadKeyPoints(): void {
      this.tourService.getKeyPointsByTour(this.tourId).subscribe({
        next: (kps) => {
          this.zone.run(() => {
            this.keyPoints = kps.sort((a, b) => a.order - b.order);
            this.cdr.detectChanges();
            if (this.map) this.renderKeyPointMarkers();
            if (this.execution?.status === 'active') {
              this.startPolling();
              // Init sidebar after keypoints loaded
              setTimeout(() => this.initSimSidebar(), 200);
            }
          });
        }
      });
    }
  
    private async initMap(): Promise<void> {
      const mapContainer = document.getElementById('execution-map');
      if (!mapContainer) {
        setTimeout(() => this.initMap(), 150);
        return;
      }
  
      this.L = await import('leaflet');
  
      const iconDefault = this.L.icon({
        iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
        iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
        shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
        iconSize: [25, 41], iconAnchor: [12, 41], popupAnchor: [1, -34], shadowSize: [41, 41]
      });
      this.L.Marker.prototype.options.icon = iconDefault;
  
      const position = this.positionService.getPosition();
      const lat = position?.latitude ?? 44.0;
      const lng = position?.longitude ?? 21.0;
  
      this.map = this.L.map('execution-map').setView([lat, lng], 13);
      this.L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors'
      }).addTo(this.map);
  
      if (position) {
        this.updatePlayerMarker(position.latitude, position.longitude);
      }
  
      if (this.keyPoints.length > 0) {
        this.renderKeyPointMarkers();
      }
    }
  
    private updatePlayerMarker(lat: number, lng: number): void {
      if (!this.map || !this.L) return;
  
      const playerIcon = this.L.divIcon({
        html: '<div style="background:#4f7cff;border:3px solid white;border-radius:50%;width:20px;height:20px;box-shadow:0 2px 8px rgba(79,124,255,0.5)"></div>',
        iconSize: [20, 20], iconAnchor: [10, 10], className: ''
      });
  
      if (this.playerMarker) {
        this.playerMarker.setLatLng([lat, lng]);
      } else {
        this.playerMarker = this.L.marker([lat, lng], { icon: playerIcon })
          .addTo(this.map).bindPopup('Vaša pozicija');
      }
    }
  
    private renderKeyPointMarkers(): void {
      if (!this.map || !this.L) return;
  
      this.keyPointMarkers.forEach(m => this.map.removeLayer(m));
      this.keyPointMarkers = [];
  
      const completedIds = new Set(
        this.execution?.completedKeyPoints?.map(ckp => ckp.keyPointId) ?? []
      );
  
      this.keyPoints.forEach(kp => {
        const isCompleted = kp.id ? completedIds.has(kp.id) : false;
  
        const icon = this.L.divIcon({
          html: `<div style="background:${isCompleted ? '#10b981' : '#f59e0b'};color:white;border-radius:50%;width:28px;height:28px;display:flex;align-items:center;justify-content:center;font-weight:bold;font-size:12px;border:2px solid white;box-shadow:0 2px 6px rgba(0,0,0,0.3)">${kp.order}</div>`,
          iconSize: [28, 28], iconAnchor: [14, 14], className: ''
        });
  
        const marker = this.L.marker([kp.latitude, kp.longitude], { icon })
          .addTo(this.map)
          .bindPopup(`<b>${kp.order}. ${kp.name}</b><br>${isCompleted ? '✅ Kompletovano' : '⏳ Nije posećeno'}`);
  
        this.keyPointMarkers.push(marker);
      });
  
      if (this.keyPoints.length > 0) {
        const bounds = this.L.latLngBounds(this.keyPoints.map(kp => [kp.latitude, kp.longitude]));
        this.map.fitBounds(bounds, { padding: [40, 40] });
      }
    }
  
    focusKeyPoint(kp: KeyPoint): void {
      if (!this.map) return;
      this.map.setView([kp.latitude, kp.longitude], 16);
    }
  
    // ---- Polling & Countdown ----
  
    private startPolling(): void {
      this.countdownSeconds = 10;
      this.pollingInterval = setInterval(() => { this.checkKeyPoint(); }, 10000);
      this.startCountdown();
    }
  
    private stopPolling(): void {
      if (this.pollingInterval) { clearInterval(this.pollingInterval); this.pollingInterval = null; }
      this.stopCountdown();
    }
  
    private startCountdown(): void {
      this.stopCountdown();
      this.countdownSeconds = 10;
      this.countdownInterval = setInterval(() => {
        this.zone.run(() => {
          this.countdownSeconds = this.countdownSeconds > 0 ? this.countdownSeconds - 1 : 10;
          this.cdr.detectChanges();
        });
      }, 1000);
    }
  
    private stopCountdown(): void {
      if (this.countdownInterval) { clearInterval(this.countdownInterval); this.countdownInterval = null; }
    }
  
    manualCheck(): void {
      this.checkingNow = true;
      this.checkKeyPoint();
      // Reset countdown
      this.startCountdown();
    }
  
    private checkKeyPoint(): void {
      if (!this.execution || this.execution.status !== 'active') {
        this.stopPolling();
        return;
      }
  
      const position = this.positionService.getPosition();
      if (!position) { this.checkingNow = false; return; }
  
      this.currentSimPosition = position;
      this.updatePlayerMarker(position.latitude, position.longitude);
  
      this.executionService.checkKeyPoint(
        this.execution.id,
        position.latitude,
        position.longitude
      ).subscribe({
        next: (result: CheckKeyPointResponse) => {
          this.zone.run(() => {
            this.checkingNow = false;
            this.execution = result.execution;
  
            if (result.keyPointReached && result.keyPoint) {
              this.lastReachedKeyPoint = result.keyPoint.name;
              this.lastCheckMessage = `✅ Dostigli ste ključnu tačku: ${result.keyPoint.name}!`;
            } else {
              this.lastCheckMessage = `📍 Provjera u ${new Date().toLocaleTimeString()} — niste blizu ključne tačke.`;
              this.lastReachedKeyPoint = null;
            }
  
            if (result.execution.status === 'completed') {
              this.stopPolling();
              this.lastCheckMessage = '🎉 Čestitamo! Kompletirali ste sve ključne tačke!';
            }
  
            this.renderKeyPointMarkers();
            this.cdr.detectChanges();
          });
        },
        error: () => {
          this.zone.run(() => {
            this.checkingNow = false;
            this.lastCheckMessage = 'Greška pri provjeri pozicije.';
            this.cdr.detectChanges();
          });
        }
      });
    }
  
    completeTour(): void {
      if (!this.execution) return;
      this.stopPolling();
  
      this.executionService.complete(this.execution.id).subscribe({
        next: (exec) => {
          this.zone.run(() => {
            this.execution = exec;
            this.lastCheckMessage = '🎉 Tura uspješno završena!';
            this.cdr.detectChanges();
          });
        },
        error: (err) => {
          this.zone.run(() => {
            this.error = err.error?.error || 'Greška pri završavanju ture.';
            this.cdr.detectChanges();
          });
        }
      });
    }
  
    abandonTour(): void {
      if (!this.execution) return;
      if (!confirm('Da li ste sigurni da želite napustiti turu?')) return;
      this.stopPolling();
  
      this.executionService.abandon(this.execution.id).subscribe({
        next: (exec) => {
          this.zone.run(() => {
            this.execution = exec;
            this.lastCheckMessage = '👋 Napustili ste turu.';
            this.cdr.detectChanges();
          });
        },
        error: (err) => {
          this.zone.run(() => {
            this.error = err.error?.error || 'Greška pri napuštanju ture.';
            this.cdr.detectChanges();
          });
        }
      });
    }
  
    getCompletedTime(kp: KeyPoint): Date | null {
      if (!kp.id || !this.execution) return null;
      const ckp = this.execution.completedKeyPoints?.find(c => c.keyPointId === kp.id);
      return ckp?.reachedAt ? new Date(ckp.reachedAt) : null;
    }
  
    get completedCount(): number { return this.execution?.completedKeyPoints?.length ?? 0; }
    get totalKeyPoints(): number { return this.keyPoints.length; }
  
    isKeyPointCompleted(kp: KeyPoint): boolean {
      if (!kp.id || !this.execution) return false;
      return this.execution.completedKeyPoints.some(ckp => ckp.keyPointId === kp.id);
    }
  
    get isActive(): boolean { return this.execution?.status === 'active'; }
  }