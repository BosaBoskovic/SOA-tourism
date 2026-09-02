import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { Router, RouterLink } from '@angular/router';
import { of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { AuthService } from '../auth/services/auth.service';
import { TourService } from '../services/tour.service';
import { ExecutionService } from '../services/execution.service';
import { CartService } from '../services/cart.service';
import { NotificationService } from '../services/notification.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent implements OnInit {
  currentUser: any = null;

  // ── Tourist widgets
  activeExecutionsCount = 0;
  purchasedToursCount = 0;

  // ── Guide widgets
  myToursCount = 0;
  publishedToursCount = 0;
  draftToursCount = 0;
  totalRevenue = 0;

  // ── Admin widgets
  totalAccounts = 0;
  blockedAccounts = 0;

  // ── Shared
  unreadNotifications = 0;

  constructor(
    private authService: AuthService,
    private router: Router,
    private cdr: ChangeDetectorRef,
    private tourService: TourService,
    private executionService: ExecutionService,
    private cartService: CartService,
    private notificationService: NotificationService,
    private http: HttpClient
  ) {}

  ngOnInit(): void {
    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;
      if (!user) {
        this.router.navigate(['/login']);
        return;
      }
      this.loadWidgets();
    });
  }

  // Puni widgete iz podataka koji već postoje u drugim servisima - dashboard
  // ne uvodi nove endpointe, samo agregira ono što je već dostupno po ulozi.
  // Svaki poziv ima catchError fallback jer je ovo prikaz na "best effort"
  // bazi - ako jedan widget ne uspije da se učita, ostali ne smiju da padnu.
  private loadWidgets(): void {
    if (!this.currentUser) return;

    this.notificationService.list().pipe(
      catchError(() => of({ notifications: [] as { read: boolean }[] }))
    ).subscribe(res => {
      this.unreadNotifications = res.notifications.filter(n => !n.read).length;
      this.cdr.detectChanges();
    });

    if (this.currentUser.role === 'tourist') {
      this.executionService.getByTourist(this.currentUser.username).pipe(
        catchError(() => of([]))
      ).subscribe(execs => {
        this.activeExecutionsCount = execs.filter(e => e.status === 'active').length;
        this.cdr.detectChanges();
      });

      this.cartService.getPurchasedTours(this.currentUser.username).pipe(
        catchError(() => of([]))
      ).subscribe(tokens => {
        this.purchasedToursCount = tokens.length;
        this.cdr.detectChanges();
      });
    } else if (this.currentUser.role === 'guide') {
      this.tourService.getToursByAuthor(this.currentUser.username).pipe(
        catchError(() => of([]))
      ).subscribe(tours => {
        this.myToursCount = tours.length;
        this.publishedToursCount = tours.filter(t => t.status === 'published').length;
        this.draftToursCount = tours.filter(t => t.status === 'draft').length;
        this.cdr.detectChanges();

        const ids = tours.map(t => t.id);
        this.tourService.getAnalytics(ids).pipe(
          catchError(() => of([]))
        ).subscribe(analytics => {
          this.totalRevenue = analytics.reduce((sum, a) => sum + a.revenue, 0);
          this.cdr.detectChanges();
        });
      });
    } else if (this.currentUser.role === 'admin') {
      this.http.get<{ accounts: any[] }>(`${environment.apiUrl}/stakeholders/accounts`).pipe(
        catchError(() => of({ accounts: [] as any[] }))
      ).subscribe(res => {
        this.totalAccounts = res.accounts.length;
        this.blockedAccounts = res.accounts.filter(a => a.isBlocked).length;
        this.cdr.detectChanges();
      });
    }
  }
}
