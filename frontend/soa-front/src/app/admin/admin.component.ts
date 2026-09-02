import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { ConfirmDialogService } from '../shared/confirm-dialog/confirm-dialog.service';
import { ToastService } from '../shared/toast/toast.service';
import { SpinnerComponent } from '../shared/ui/spinner/spinner.component';
import { EmptyStateComponent } from '../shared/ui/empty-state/empty-state.component';

@Component({
  selector: 'app-admin',
  standalone: true,
  imports: [CommonModule, FormsModule, SpinnerComponent, EmptyStateComponent],
  templateUrl: './admin.component.html',
  styleUrl: './admin.component.css'
})
export class AdminComponent implements OnInit {
  accounts: any[] = [];
  loading = false;
  errorMessage = '';

  // Pretraga i paginacija rade nad već učitanom listom - GET /accounts nema
  // svoj query-param filter na backendu, a lista naloga u ovoj demo aplikaciji
  // je dovoljno mala da klijentska paginacija ima smisla.
  searchTerm = '';
  page = 0;
  pageSize = 10;

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

  get filteredAccounts(): any[] {
    const term = this.searchTerm.trim().toLowerCase();
    if (!term) return this.accounts;
    return this.accounts.filter(a =>
      a.username?.toLowerCase().includes(term) || a.email?.toLowerCase().includes(term)
    );
  }

  get totalPages(): number {
    return Math.max(1, Math.ceil(this.filteredAccounts.length / this.pageSize));
  }

  get pagedAccounts(): any[] {
    const start = this.page * this.pageSize;
    return this.filteredAccounts.slice(start, start + this.pageSize);
  }

  onSearchChange(): void {
    this.page = 0; // rezultat pretrage počinje uvijek od prve strane
  }

  goToPage(page: number): void {
    if (page < 0 || page >= this.totalPages) return;
    this.page = page;
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
        this.toastService.error('Greška pri blokiranju naloga.');
      }
    });
  }

  async unblockAccount(account: any): Promise<void> {
    const confirmed = await this.confirmDialogService.confirm(
      `Da li ste sigurni da želite da odblokirate nalog "${account.username}"?`
    );
    if (!confirmed) return;

    this.http.patch(`${this.apiUrl}/accounts/${account.username}/unblock`, {}).subscribe({
      next: () => {
        account.isBlocked = false;
        this.cdr.detectChanges();
        this.toastService.success(`Nalog "${account.username}" je odblokiran.`);
      },
      error: () => {
        this.toastService.error('Greška pri odblokiranju naloga.');
      }
    });
  }

  async deleteAccount(account: any): Promise<void> {
    const confirmed = await this.confirmDialogService.confirm(
      `Da li sigurno želite da trajno obrišete nalog "${account.username}"? Ova radnja se ne može poništiti.`,
      { confirmLabel: 'Obriši' }
    );
    if (!confirmed) return;

    this.http.delete(`${this.apiUrl}/accounts/${account.username}`).subscribe({
      next: () => {
        this.accounts = this.accounts.filter(a => a.username !== account.username);
        this.toastService.success(`Nalog "${account.username}" je obrisan.`);
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.toastService.error(err.error?.error || 'Greška pri brisanju naloga.');
      }
    });
  }
}
