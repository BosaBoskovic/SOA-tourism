import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { forkJoin, of, Observable } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { ExecutionService, TourExecution } from '../../services/execution.service';
import { TourService, Tour } from '../../services/tour.service';
import { AuthService } from '../../auth/services/auth.service';

@Component({
  selector: 'app-my-executions',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './my-executions.component.html',
  styleUrl: './my-executions.component.css'
})
export class MyExecutionsComponent implements OnInit {
  executions: TourExecution[] = [];
  tourNames: Record<string, string> = {};
  loading = false;
  error = '';

  constructor(
    private executionService: ExecutionService,
    private tourService: TourService,
    private authService: AuthService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit(): void {
    const user = this.authService.getCurrentUser();
    if (!user) return;

    this.loading = true;
    this.executionService.getByTourist(user.username).subscribe({
      next: (executions) => {
        this.executions = executions.sort((a, b) => b.startedAt.localeCompare(a.startedAt));
        this.loadTourNames();
      },
      error: () => {
        this.error = 'Greška pri učitavanju tura.';
        this.loading = false;
        this.cdr.detectChanges();
      }
    });
  }

  private loadTourNames(): void {
    const uniqueTourIds = Array.from(new Set(this.executions.map(e => e.tourId)));
    if (uniqueTourIds.length === 0) {
      this.loading = false;
      this.cdr.detectChanges();
      return;
    }

    const requests: Record<string, Observable<Tour | null>> = {};
    for (const id of uniqueTourIds) {
      requests[id] = this.tourService.getTourById(id).pipe(catchError(() => of(null)));
    }

    forkJoin(requests).subscribe((results: Record<string, Tour | null>) => {
      for (const id of uniqueTourIds) {
        this.tourNames[id] = results[id]?.name || 'Nepoznata tura';
      }
      this.loading = false;
      this.cdr.detectChanges();
    });
  }

  statusLabel(status: string): string {
    switch (status) {
      case 'active': return 'U toku';
      case 'completed': return 'Završena';
      case 'abandoned': return 'Napuštena';
      default: return status;
    }
  }
}
