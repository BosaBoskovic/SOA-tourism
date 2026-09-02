import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';

export interface ConfirmRequest {
  message: string;
  confirmLabel: string;
  cancelLabel: string;
  resolve: (confirmed: boolean) => void;
}

export interface ConfirmOptions {
  confirmLabel?: string;
  cancelLabel?: string;
}

// Promise-based replacement for the native confirm() calls (or missing
// confirmation entirely) on destructive actions - await this instead.
@Injectable({ providedIn: 'root' })
export class ConfirmDialogService {
  private readonly requestSubject = new Subject<ConfirmRequest>();
  readonly requests$ = this.requestSubject.asObservable();

  confirm(message: string, options: ConfirmOptions = {}): Promise<boolean> {
    return new Promise(resolve => {
      this.requestSubject.next({
        message,
        confirmLabel: options.confirmLabel ?? 'Potvrdi',
        cancelLabel: options.cancelLabel ?? 'Otkaži',
        resolve,
      });
    });
  }
}
