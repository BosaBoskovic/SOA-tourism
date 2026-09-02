import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

@Component({
  selector: 'app-admin',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './admin.component.html',
  styleUrl: './admin.component.css'
})
export class AdminComponent implements OnInit {
  accounts: any[] = [];
  loading = false;
  errorMessage = '';

  // Auth header comes from the global authInterceptorFn - no need to attach it per call here.
  private apiUrl = `${environment.apiUrl}/stakeholders`;

  constructor(
    private http: HttpClient,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit(): void {
    this.loadAccounts();
  }

  loadAccounts(): void {
    this.loading = true;
    this.errorMessage = '';

    this.http.get<{ accounts: any[] }>(`${this.apiUrl}/accounts`).subscribe({
      next: response => {
        this.accounts = response.accounts ?? [];
        this.loading = false;
        this.cdr.detectChanges();
      },
      error: () => {
        this.errorMessage = 'Greška pri učitavanju naloga.';
        this.loading = false;
        this.cdr.detectChanges();
      }
    });
  }

  blockAccount(account: any): void {
    this.http.patch(`${this.apiUrl}/accounts/${account.username}/block`, {}).subscribe({
      next: () => {
        account.isBlocked = true;
        this.cdr.detectChanges();
      },
      error: () => {
        this.errorMessage = 'Greška pri blokiranju naloga.';
        this.cdr.detectChanges();
      }
    });
  }
}
