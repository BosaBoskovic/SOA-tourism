import { Component, OnDestroy, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subscription, interval, startWith, switchMap } from 'rxjs';
import { NotificationService, AppNotification } from '../../services/notification.service';
import { AuthService } from '../../auth/services/auth.service';

const POLL_INTERVAL_MS = 30000;

@Component({
  selector: 'app-notification-bell',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './notification-bell.component.html',
  styleUrl: './notification-bell.component.css'
})
export class NotificationBellComponent implements OnInit, OnDestroy {
  notifications: AppNotification[] = [];
  open = false;
  private pollSub?: Subscription;

  constructor(
    private notificationService: NotificationService,
    private authService: AuthService,
    private cdr: ChangeDetectorRef
  ) {}

  get unreadCount(): number {
    return this.notifications.filter(n => !n.read).length;
  }

  ngOnInit(): void {
    this.pollSub = this.authService.currentUser$.pipe(
      switchMap(user => {
        if (!user) return [];
        return interval(POLL_INTERVAL_MS).pipe(startWith(0));
      }),
      switchMap(() => this.notificationService.list())
    ).subscribe({
      next: (res) => {
        this.notifications = res.notifications;
        this.cdr.detectChanges();
      },
      error: () => {
        // Silent - a failed poll just means the badge doesn't update this cycle.
      }
    });
  }

  ngOnDestroy(): void {
    this.pollSub?.unsubscribe();
  }

  toggle(): void {
    this.open = !this.open;
  }

  markRead(n: AppNotification): void {
    if (n.read) return;
    n.read = true;
    this.notificationService.markRead(n.id).subscribe();
  }

  markAllRead(): void {
    this.notifications.forEach(n => (n.read = true));
    this.notificationService.markAllRead().subscribe();
  }
}
