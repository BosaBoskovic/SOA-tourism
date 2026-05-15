import { Component, OnInit, NgZone, ChangeDetectorRef  } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient, HttpHeaders } from '@angular/common/http';


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

  private apiUrl = 'http://localhost:8080/stakeholders';

 constructor(
  private http: HttpClient,
  private zone: NgZone,
   private cdr: ChangeDetectorRef
) {}

  ngOnInit(): void {
    this.loadAccounts();
  }

loadAccounts(): void {
  console.log('Ucitavam naloge...');
  this.loading = true;
  this.errorMessage = '';

  this.http.get<{ accounts: any[] }>(`${this.apiUrl}/accounts`, {
    headers: this.getAuthHeaders()
  }).subscribe({
    next: response => {
      console.log('Accounts response:', response);

      this.accounts = response.accounts ?? [];
      this.loading = false;

      console.log('Accounts posle setovanja:', this.accounts);
      this.cdr.detectChanges();
    },
    error: err => {
      console.log('Accounts error:', err);
      this.errorMessage = 'Greška pri učitavanju naloga.';
      this.loading = false;
      this.cdr.detectChanges();
    }
  });
}
blockAccount(account: any): void {
  this.http.patch(`${this.apiUrl}/accounts/${account.username}/block`, {}, {
    headers: this.getAuthHeaders()
  }).subscribe({
    next: () => {
      account.isBlocked = true;
    },
    error: err => {
      console.log('Block error:', err);
      this.errorMessage = 'Greška pri blokiranju naloga.';
    }
  });
}
  private getAuthHeaders(): HttpHeaders {
    const token = localStorage.getItem('token');
    return new HttpHeaders({
      Authorization: `Bearer ${token}`
    });
  }
}