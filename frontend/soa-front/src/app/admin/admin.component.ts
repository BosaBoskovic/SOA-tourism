import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { ConfirmDialogService } from '../shared/confirm-dialog/confirm-dialog.service';
import { ToastService } from '../shared/toast/toast.service';

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
    private cdr: ChangeDetectorRef,
    private confirmDialogService: ConfirmDialogService,
    private toastService: ToastService
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

  async blockAccount(account: any): Promise<void> {
    const confirmed = await this.confirmDialogService.confirm(
      `Da li ste sigurni da želite da blokirate nalog "${account.username}"?`
    );
    if (!confirmed) return;

    this.http.patch(`${this.apiUrl}/accounts/${account.username}/block`, {}).subscribe({
      next: () => {
        account.isBlocked = true;
        this.cdr.detectChanges();
        this.toastService.success(`Nalog "${account.username}" je blokiran.`);
      },
      error: () => {
        this.errorMessage = 'Greška pri blokiranju naloga.';
        this.cdr.detectChanges();
        this.toastService.error(this.errorMessage);
      }
    });
  }
}
