import { Component, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subscription } from 'rxjs';
import { ConfirmDialogService, ConfirmRequest } from './confirm-dialog.service';

@Component({
  selector: 'app-confirm-dialog',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './confirm-dialog.component.html',
  styleUrl: './confirm-dialog.component.css'
})
export class ConfirmDialogComponent implements OnDestroy {
  active: ConfirmRequest | null = null;
  private sub: Subscription;

  constructor(private confirmDialogService: ConfirmDialogService) {
    this.sub = this.confirmDialogService.requests$.subscribe(req => {
      this.active = req;
    });
  }

  respond(confirmed: boolean): void {
    this.active?.resolve(confirmed);
    this.active = null;
  }

  ngOnDestroy(): void {
    this.sub.unsubscribe();
  }
}
